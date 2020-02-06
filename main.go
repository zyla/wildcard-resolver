package main

import (
	"fmt"
	"net"
	"context"
    "os"
    "strings"

	"github.com/miekg/dns"
    "github.com/ryanuber/go-glob"
)

type resolver struct {
    mappings []mapping
}

type mapping struct {
    pattern string
    target string
}

func main() {
	var r resolver

    config := os.Getenv("HOSTS")
    if len(config) > 0 {
        pairs := strings.Split(config, ",")
        for _, pair := range pairs {
            split := strings.Split(pair, ":")
            r.mappings = append(r.mappings, mapping{
                pattern: split[0],
                target: split[1],
            })
        }
    }

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

	origname := query.Question[0].Name
    name := strings.Trim(origname, ".")
	switch query.Question[0].Qtype {
	case dns.TypeA:
        var target string

        for _, m := range r.mappings {
            if glob.Glob(m.pattern, name) {
                target = m.target
                break
            }
        }

        fmt.Printf("Serving %v=%v\n", name, target)
        ips, err := net.DefaultResolver.LookupIPAddr(context.Background(), target)

        if err == nil {
            fmt.Printf("%v=%v is %v\n", name, target, ips[0].IP)
            resp = new(dns.Msg)
            rr := &dns.A{
                Hdr: dns.RR_Header{Name: origname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
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
