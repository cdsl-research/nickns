package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	. "nickns/resolver"

	// "github.com/cybozu-go/well"
	"github.com/miekg/dns"
)

var Port = 5300
var Domains = []string{"local.", "a910.tak-cslab.org."}

func stripDomainName(fqdn string) string {
	for _, domain := range Domains {
		r := regexp.MustCompile("^[0-9a-zA-Z_\\-]+." + domain + "$")
		if r.MatchString(fqdn) {
			// log.Println(strings.Replace(fqdn, domain, "", -1))
			return strings.Replace(fqdn, "."+domain, "", -1)
		}
	}
	return fqdn
}

func dnsRequestHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// normal request
	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				hostname := stripDomainName(q.Name)
				if ip := ResolveRecordTypeA(hostname); ip != "" {
					log.Printf("[QueryHit] %s => %s\n", q.Name, ip)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			case dns.TypePTR:
				if hostname := ResolveRecordTypePTR(q.Name); hostname != "" {
					for _, domain := range Domains {
						fqdn := hostname + "." + domain
						log.Printf("[QueryHit] %s => %s\n", q.Name, fqdn)
						rr, err := dns.NewRR(fmt.Sprintf("%s PTR %s", q.Name, fqdn))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
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
	for _, domain := range Domains {
		dns.HandleFunc(domain, dnsRequestHandler)
	}
	dns.HandleFunc("arpa.", dnsRequestHandler)

	// bootstrapping dns server
	port := 5300
	server := &dns.Server{Addr: ":" + strconv.Itoa(Port), Net: "udp"}
	log.Printf("NickNS Starting at %d/udp\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
	defer server.Shutdown()
}
