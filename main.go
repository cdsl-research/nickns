package main

import (
	"flag"
	"fmt"
	"github.com/cdsl-research/nickns/lib"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

type configOptions struct {
	Port    int
	TTL     int
	Domains []string
}

var (
	confPathOpt = flag.String("c", "config.toml", "Path to config.toml")
	hostPathOpt = flag.String("n", "hosts.toml", "Path to hosts.toml")
)

var confOptions = configOptions{
	Port:    5300,
	TTL:     3600,
	Domains: []string{"local.", "example.com."},
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
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
				if ip := lib.ResolveRecordTypeA(hostname); ip != "" {
					log.Printf("[QueryHit] %s => %s\n", q.Name, ip)
					rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", q.Name, confOptions.TTL, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			case dns.TypePTR:
				if hostname := lib.ResolveRecordTypePTR(q.Name); hostname != "" {
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
	if !fileExists(*confPathOpt) {
		log.Fatalln("Fail to load ", *confPathOpt)
	}
	content, err := ioutil.ReadFile(*confPathOpt)
	if err != nil {
		log.Fatalln("Fail to read conf file: ", err)
	}
	if _, err := toml.Decode(string(content), &confOptions); err != nil {
		log.Fatalln("Fail to decode conf as toml: ", err)
	}

	// set path hosts.toml
	if !fileExists(*hostPathOpt) {
		log.Fatalln("Could not load ", *hostPathOpt)
	}
	lib.SetEsxiConfigPath(*hostPathOpt)

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
