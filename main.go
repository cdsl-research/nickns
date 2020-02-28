package main

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	// "github.com/cybozu-go/well"
	"github.com/miekg/dns"
)

// todo: only support 'A' record
type QueryCache struct {
	Fqdn   string
  IpAddr string
	Expire time.Time
}

type Machine struct {
	Id   int
	Name string
}

type Machines []Machine

func sshGetAllVms(ip string, port string, config *ssh.ClientConfig) (bytes.Buffer, error) {
	var buf bytes.Buffer

	conn, err := ssh.Dial("tcp", ip+":"+port, config)
	if err != nil {
		return buf, err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return buf, err
	}
	defer session.Close()

	session.Stdout = &buf
	remote_command := "cat /tmp/result"
	// remote_command := "vim-cmd vmsvc/getallvms"
	if err := session.Run(remote_command); err != nil {
		return buf, err
	}

	return buf, nil
}

func parseResult(buf bytes.Buffer) Machines {
	r := regexp.MustCompile(`^\d.+`)
	var vms Machines
	for {
		st, err := buf.ReadString('\n')
		if err != nil {
			return vms
		}

		if !r.MatchString(st) {
			continue
		}

		slice := strings.Split(st, "    ")
		slice0, err := strconv.Atoi(slice[0])
		if err != nil {
			slice0 = -1
		}
		vms = append(vms, Machine{
			Id:   slice0,
			Name: strings.TrimSpace(slice[1]),
		})
	}
}

func getVmIp(ip string, port string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", ip+":"+port, config)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf
	remote_command := "echo 192.168.100.100"
	// remote_command := "vim-cmd vmsvc/get.summary 16 | grep ipAddress | grep -o [0-9\.]\\+"
	if err := session.Run(remote_command); err != nil {
		log.Println(err.Error())
		return ""
	}
	return buf.String()
}

func resolveRecordTypeA(fqdn string) string {
	// ssh connect
	ip := "127.0.0.1"
	port := "2200"
	user := "root"
	pass := "THEPASSWORDYOUCREATED"
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
	}

	b, err := sshGetAllVms(ip, port, config)
	if err != nil {
		log.Println(err.Error())
	}

	vms := parseResult(b)
	for _, vm := range vms {
		// debug:: println(vm.Name, "and", fqdn)
		if vm.Name == strings.Split(fqdn, ".")[0] {
			// return "192.168.0.1" // Hit
			return getVmIp(ip, port, config) // Hit
		}
	}
	return "" // UnHit
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			ip := resolveRecordTypeA(q.Name)
			if ip != "" {
				log.Printf("[Hit]\tQuery for %s\n", q.Name)
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else {
				log.Printf("[UnHit]\tQuery for %s\n", q.Name)
			}
		}
	}
}

func dnsRequestHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	// attach request handler func
	dns.HandleFunc("local.", dnsRequestHandler)

	// dns server
	port := 5300
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
