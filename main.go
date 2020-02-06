package main

import (
	"fmt"
	"net"
	"context"

	"github.com/miekg/dns"
)

// resolver implements the Resolver interface
type resolver struct {
}

func main() {

	var r resolver

	addr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 53,
	}

    conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		err = fmt.Errorf("error in opening name server socket %v", err)
		return
	}
	s := &dns.Server{Handler: &r, PacketConn: conn}
	s.ActivateAndServe()
}

func (r *resolver) ServeDNS(w dns.ResponseWriter, query *dns.Msg) {
    var resp *dns.Msg

	if query == nil || len(query.Question) == 0 {
		return
	}

	name := query.Question[0].Name
	switch query.Question[0].Qtype {
	case dns.TypeA:
        fmt.Printf("Serving %v\n", name)
        ips, err := net.DefaultResolver.LookupIPAddr(context.Background(), name)

        if err == nil {
            resp = new(dns.Msg)
            fmt.Printf("Success %v\n", name)
            rr := &dns.A{
                Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
                A:   ips[0].IP,
            }
            resp.SetReply(query)
            resp.SetRcode(query, dns.RcodeSuccess)
			resp.Answer = append(resp.Answer, rr)
		} else {
            resp = new(dns.Msg)
            resp.SetRcode(query, dns.RcodeServerFailure)
        }
	}

	if resp == nil {
        fmt.Printf("Fail %v\n", name)
		resp = new(dns.Msg)
		resp.SetRcode(query, dns.RcodeServerFailure)
	}
    w.WriteMsg(resp)
}
