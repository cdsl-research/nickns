package resolver

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	. "nickns/resolver/esxi"

	"golang.org/x/crypto/ssh"
)

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

// Resolve type 'A' record
func ResolveRecordTypeA(hostname string) string {
	// cache hit
	/*
		if hasCache[fqdn] {
			log.Printf("[CacheHit] %s\n", fqdn)
			for _, vm := range cache {
				// todo: check cache-expire -> del cache from set and array
				if vm.Fqdn == strings.Split(fqdn, ".")[0] {
					return vm.IpAddr
				}
			}
		}
	*/
	for _, vm := range GetAllVmIdName() {
		// log.Println(vm.Name, fqdn)
		if vm.Name == hostname {
			return GetVmIp(vm)
		}
	}
	return "" // UnHit
}

// Resolve type 'A' record
func ResolveRecordTypePTR(ptrAddr string) string {
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

	b, err := ExecCommandSsh(ip, port, config, `
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
		vmFqdn := strings.Replace(slice[1], "\n", "", -1)
		hasCache[vmFqdn+".local."] = true
		cache = append(cache, QueryCache{
			Fqdn:   vmFqdn,
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
