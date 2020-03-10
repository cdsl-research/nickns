package esxi

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

// For command result
type Machine struct {
	// VM ID on ESXi
	Id int
	// VM Name on ESXi
	Name string
}
type Machines []Machine

// Parse command result on SshGetAllVms()
func ParseResultAllVms(buf bytes.Buffer) Machines {
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

// Get VM info from ESXi via SSH
func ExecCommandSsh(ip string, port string, config *ssh.ClientConfig, command string) (bytes.Buffer, error) {
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
func GetVmIp(ip string, port string, config *ssh.ClientConfig, vmid int) string {
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