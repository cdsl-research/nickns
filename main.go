package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"flag"

	. "nickns/resolver"

	// "github.com/cybozu-go/well"
	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

type configOptions struct {
	Port    int
	TTL     int
	Domains []string
}

var (
	portOpt = flag.Int("p", 5300, "Listening udp port")
	ttlOpt = flag.Int("t", 3600, "Interval for keeping DNS cache")
	confPathOpt = flag.String("c", "config.toml", "Path to config.toml")
	hostPathOpt = flag.String("n", "hosts.toml", "Path to hosts.toml")
)

var confOptions = configOptions{
	Port:    *portOpt,
	TTL:     *ttlOpt,
	Domains: []string{"local.", "example.com."},
}

func stripDomainName(fqdn string) string {
	for _, domain := range confOptions.Domains {
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
					rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", q.Name, confOptions.TTL, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			case dns.TypePTR:
				if hostname := ResolveRecordTypePTR(q.Name); hostname != "" {
					for _, domain := range confOptions.Domains {
						fqdn := hostname + "." + domain
						log.Printf("[QueryHit] %s => %s\n", q.Name, fqdn)
						rr, err := dns.NewRR(fmt.Sprintf("%s %d IN PTR %s", q.Name, confOptions.TTL, fqdn))
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
	// parse command line params
	flag.Parse()

	// load config
	content, err := ioutil.ReadFile(*confPathOpt)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := toml.Decode(string(content), &confPathOpt); err != nil {
		log.Fatalln(err)
	}

	// attach request handler func
	for _, domain := range confOptions.Domains {
		dns.HandleFunc(domain, dnsRequestHandler)
	}
	dns.HandleFunc("arpa.", dnsRequestHandler)

	// bootstrapping dns server
	server := &dns.Server{Addr: ":" + strconv.Itoa(confOptions.Port), Net: "udp"}
	log.Printf("NickNS Starting at %d/udp\n", confOptions.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
	defer server.Shutdown()
}
