package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Machine struct {
	Id         int
	Name       string
}

type Machines []Machine

func ssh_cmd(ip string, port string, config *ssh.ClientConfig) (bytes.Buffer, error) {
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
	if err := session.Run(remote_command); err != nil {
		return buf, err
	}

	return buf, nil
}

func parse_result(buf bytes.Buffer) Machines {
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

func main() {
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

	b, err := ssh_cmd(ip, port, config)
	if err != nil {
		log.Fatal(err)
	}

	vms := parse_result(b)
	fmt.Println(vms)
}
