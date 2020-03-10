package main

import (
	"fmt"
	"log"
	"strconv"

	. "nickns/solver"

	// "github.com/cybozu-go/well"
	"github.com/miekg/dns"
)

func dnsRequestHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// normal request
	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				if ip := ResolveRecordTypeA(q.Name); ip != "" {
					log.Printf("[QueryHit] %s => %s\n", q.Name, ip)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			case dns.TypePTR:
				if fqdn := ResolveRecordTypePTR(q.Name); fqdn != "" {
					log.Printf("[QueryHit] %s => %s\n", q.Name, fqdn)
					rr, err := dns.NewRR(fmt.Sprintf("%s PTR %s", q.Name, fqdn))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			}
		}
	}

	w.WriteMsg(m)
}

func main() {
	// attach request handler func
	domains := []string{"local."}
	for _,domain := range domains {
		dns.HandleFunc(domain, dnsRequestHandler)
	}
	dns.HandleFunc("arpa.", dnsRequestHandler)

	// bootstrapping dns server
	port := 5300
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("NickNS Starting at %d/udp\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
	defer server.Shutdown()
}
