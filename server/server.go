package server

import (
    "log"
    "github.com/miekg/dns"
    "github.com/apalrd/styx46/config"
    "github.com/apalrd/styx46/resolver"
    "github.com/apalrd/styx46/translator"
)

type Server struct {
    config *config.Config,
    resolver *resolver.Resolver,
    translator *translator.Translator
}

func New(cfg *config.Config,res *resolver.Resolver,xlate *translator.Translator) *Server {
    return &Server{config: cfg, resolver: res, translator: xlate}
}

func (s *Server) handleRequest(w dns.ResponseWriter, req *dns.Msg) {
    //If the request is a type A
        // kick off an AAAA request async

    //Always forward the request upstream regardless of what it is
    resp, err := s.resolver.Forward(req)

    //if we did an AAAA request async
        //wait for it


    //If original type was A and AAAA succeeded and (A failed  config.PreferSynth)
        //if config.Debug, print that we going to are synthesize query name, result ipv6 result
        //call s.translator.Lookup(net.IP) with ipv6 result to get translated ipv4
        //if config.Debug, print the synthesized result address
        //return synthesized address as A record
    //if query failed to resolve
        //if config.debug, print the query and servfail
        //return servfail
    //Else
        //if config.Debug print the query type, name, and result
        //return response we got from upstream
}

func (s *Server) ListenAndServe() error {
    mux := dns.NewServeMux()
    mux.HandleFunc(".", s.handleRequest)

    server := &dns.Server{
        Addr:    s.addr,
        Net:     "udp",
        Handler: mux,
    }

    log.Printf("DNS server listening on %s", s.addr)
    return server.ListenAndServe()
}