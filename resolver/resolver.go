package resolver

import (
    "fmt"
    "github.com/miekg/dns"
)

type Resolver struct {
    upstreams []string
    client    *dns.Client
}

func New(upstreams []string) *Resolver {
    return &Resolver{
        upstreams: upstreams,
        client:    &dns.Client{},
    }
}

// Forward forwards a DNS message to the first responding upstream.
func (r *Resolver) Forward(req *dns.Msg) (*dns.Msg, error) {
    for _, upstream := range r.upstreams {
        resp, _, err := r.client.Exchange(req, upstream)
        if err == nil {
            return resp, nil
        }
    }
    return nil, fmt.Errorf("all upstreams failed")
}