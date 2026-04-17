package server

import (
    "log"
    "github.com/miekg/dns"
    "github.com/apalrd/styx46/resolver"
)

type Server struct {
    addr     string
    resolver *resolver.Resolver
}

func New(addr string, res *resolver.Resolver) *Server {
    return &Server{addr: addr, resolver: res}
}

func (s *Server) handleRequest(w dns.ResponseWriter, req *dns.Msg) {
    resp, err := s.resolver.Forward(req)
    if err != nil {
        log.Printf("resolve error: %v", err)
        m := new(dns.Msg)
        m.SetRcode(req, dns.RcodeServerFailure)
        w.WriteMsg(m)
        return
    }
    w.WriteMsg(resp)
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