package server

import (
    "log"
    "net"
    "github.com/miekg/dns"
    "github.com/apalrd/styx46/config"
    "github.com/apalrd/styx46/resolver"
    "github.com/apalrd/styx46/translator"
)

type Server struct {
    config *config.Config
    resolver *resolver.Resolver
    translator *translator.Translator
}

func New(cfg *config.Config,res *resolver.Resolver,xlate *translator.Translator) *Server {
    return &Server{config: cfg, resolver: res, translator: xlate}
}

func (s *Server) handleRequest(w dns.ResponseWriter, req *dns.Msg) {
    //vars we will need async
	var (
		aaaaResp *dns.Msg
		aaaaErr  error
		doneCh   chan struct{}
	)

    //orignal question is A?
	q := req.Question[0]
	isA := q.Qtype == dns.TypeA

	// If A request → kick off AAAA asynchronously
	if isA {
		doneCh = make(chan struct{})
		go func() {
			defer close(doneCh)

			aaaaReq := new(dns.Msg)
			aaaaReq.SetQuestion(q.Name, dns.TypeAAAA)

			aaaaResp, aaaaErr = s.resolver.Forward(aaaaReq)
		}()
	}

    //Always forward the request upstream regardless of what it is
    resp, err := s.resolver.Forward(req)

	// wait for AAAA if we launched it
	if isA && doneCh != nil {
		<-doneCh
	}

	// check if response failed (no answers or error)
	aFailed := err != nil || resp == nil || len(resp.Answer) == 0
    aaaaFailed := isA && aaaaErr != nil || aaaaResp == nil || len(aaaaResp.Answer) == 0

    // get the ipv6 response (first one, if there are multiple)
	var ipv6 net.IP
    if(!aaaaFailed) {
        for _, ans := range aaaaResp.Answer {
            if aaaa, ok := ans.(*dns.AAAA); ok {
                ipv6 = aaaa.AAAA
                break
            }
		}
    }


    //If original type was A and AAAA succeeded (do synthesis)
	if isA && ipv6 != nil && (aFailed || s.config.PreferSynth) {
		if s.config.Debug {
			log.Printf("Synthesizing A for %s from AAAA %s", q.Name, ipv6.String())
		}

        //ask the translator to translate this v6 address
		ipv4, ipv4err := s.translator.Lookup(ipv6)
		if ipv4err == nil && ipv4 != nil {
            //translator did not return an error
			if s.config.Debug {
				log.Printf("Synthesized IPv4: %s", ipv4.String())
			}

            //return synthesized A result
			msg := new(dns.Msg)
			msg.SetReply(req)

			a := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: ipv4,
			}

			msg.Answer = []dns.RR{a}
			_ = w.WriteMsg(msg)
			return
		}
    }

    // If original query failed
	if err != nil || resp == nil {
		if s.config.Debug {
			log.Printf("Query failed: %s (%d) → SERVFAIL", q.Name, q.Qtype)
		}

		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeServerFailure)
		_ = w.WriteMsg(m)
		return
	}

    // Else, return the query we got 	
    if s.config.Debug {
		log.Printf("Query: %s (%d) → %d answers", q.Name, q.Qtype, len(resp.Answer))
	}

	_ = w.WriteMsg(resp)
}

func (s *Server) ListenAndServe() error {
    mux := dns.NewServeMux()
    mux.HandleFunc(".", s.handleRequest)

    server := &dns.Server{
        Addr:    s.config.Listen,
        Net:     "udp",
        Handler: mux,
    }

    log.Printf("DNS server listening on %s", s.config.Listen)
    return server.ListenAndServe()
}