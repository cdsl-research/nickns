package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	// "github.com/cybozu-go/well"
	"github.com/miekg/dns"
)

// For command result
type Machine struct {
	// VM ID on ESXi
	Id int
	// VM Name on ESXi
	Name string
}
type Machines []Machine

// Manage cache existing
var cache QueryCaches
var hasCache = map[string]bool{}

// Manage cache detail
type QueryCache struct {
	// todo: only support 'A' record
	Fqdn   string
	IpAddr string
	Expire time.Time
}
type QueryCaches []QueryCache

// Get VM info from ESXi via SSH
func execCommandSsh(ip string, port string, config *ssh.ClientConfig, command string) (bytes.Buffer, error) {
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
	if err := session.Run(command); err != nil {
		return buf, err
	}

	return buf, nil
}

// Resolve VM ID to VM IPAddr
func getVmIp(ip string, port string, config *ssh.ClientConfig, vmid int) string {
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
	// remoteCommand := "echo 192.168.100.100"
	remoteCommand := fmt.Sprintf("vim-cmd vmsvc/get.summary %d | grep ipAddress | grep -o [0-9\\.]\\\\+", vmid)
	if err := session.Run(remoteCommand); err != nil {
		log.Println(err.Error())
		return ""
	}
	return strings.Replace(buf.String(), "\n", "", -1)
}

// Parse command result on SshGetAllVms()
func parseResultAllVms(buf bytes.Buffer) Machines {
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

// Resolve type 'A' record
func resolveRecordTypeA(fqdn string) string {
	// cache hit
	if hasCache[fqdn] {
		log.Printf("[CacheHit] %s\n", fqdn)
		for _, vm := range cache {
			// todo: check cache-expire -> del cache from set and array
			if vm.Fqdn == strings.Split(fqdn, ".")[0] {
				return vm.IpAddr
			}
		}
	}

	// ssh key
	buf, err := ioutil.ReadFile("./old/id_rsa")
	if err != nil {
		panic(err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}

	// ssh connect
	ip := "192.168.0.20"
	port := "22"
	user := "root"
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	// get vm list via ssh
	b, err := execCommandSsh(ip, port, config, "vim-cmd vmsvc/getallvms") // test "cat /tmp/result"
	if err != nil {
		log.Println(err.Error())
	}

	// check a matched host in command result
	for _, vm := range parseResultAllVms(b) {
		// debug:: println(vm.Name, "and", fqdn)
		if vm.Name == strings.Split(fqdn, ".")[0] { // Hit
			// resolve vmid to ip
			vmIp := getVmIp(ip, port, config, vm.Id)

			// add cache
			hasCache[fqdn] = true
			cache = append(cache, QueryCache{
				Fqdn:   fqdn,
				IpAddr: vmIp,
				Expire: time.Now(),
			})

			// return "192.168.0.1"
			return vmIp
		}
	}
	return "" // UnHit
}

// Resolve type 'A' record
func resolveRecordTypePTR(ptrAddr string) string {
	// ssh key
	buf, err := ioutil.ReadFile("./old/id_rsa")
	if err != nil {
		panic(err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}

	// ssh connect
	ip := "192.168.0.20"
	port := "22"
	user := "root"
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	b, err := execCommandSsh(ip, port, config, `
		for i in $(vim-cmd vmsvc/getallvms | awk '{print $1}' | grep [0-9]\\+)
		do
			vim-cmd vmsvc/get.summary $i | egrep "\s+(name|ipAddress)" | grep -o '".*"' \
			| sed -e ':a;N;$!ba;s/\n/ /g;s/"//g' | grep [0-9\.]\\{7,\\} &
		done
	`)
	if err != nil {
		log.Println(err.Error())
	}
	// println(b.String())

	foundAnswer := ""
	for {
		st, err := b.ReadString('\n')
		if err != nil {
			break
		}

		slice := strings.Split(st, " ")
		vmIp := slice[0]
		vmFqdn := strings.Replace(slice[1],"\n", "", -1)
	    hasCache[vmFqdn+".local."] = true
		cache = append(cache, QueryCache{
			Fqdn: vmFqdn,
			IpAddr: vmIp,
			Expire: time.Now(),
		})

		slice = strings.Split(ptrAddr, ".")
		if fmt.Sprintf("%s.%s.%s.%s", slice[3], slice[2], slice[1], slice[0]) == vmIp {
			foundAnswer = vmFqdn
		}
	}

	// log.Printf("debug %s => %s\n", foundAnswer, ptrAddr)
	if foundAnswer == "" {
		return foundAnswer
	}
	return foundAnswer + ".local" + "."
}

func dnsRequestHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// parse query
	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				if ip := resolveRecordTypeA(q.Name); ip != "" {
					log.Printf("[QueryHit] %s => %s\n", q.Name, ip)
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				} else {
					log.Printf("[QueryUnHit] %s\n", q.Name)
				}
			case dns.TypePTR:
				if fqdn := resolveRecordTypePTR(q.Name); fqdn != "" {
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
