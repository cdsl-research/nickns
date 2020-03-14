package esxi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

type Machine struct {
	Id       int    // VM ID on ESXi
	Name     string // VM Name on ESXi
	NodeName string // ESXi Node Name
}
type Machines []Machine

type esxiNode struct {
	Address      string
	Port         string
	User         string
	IdentityFile string `toml:"identity_file"`
}
type esxiNodes map[string]esxiNode

// Get SSH Nodes from hosts.toml
func loadAllEsxiNodes() esxiNodes {
	content, err := ioutil.ReadFile("hosts.toml")
	if err != nil {
		log.Fatalln(err)
	}

	var nodes esxiNodes
	if _, err := toml.Decode(string(content), &nodes); err != nil {
		log.Fatalln(err)
	}
	/* Debug
	for key,value := range esxiNodes {
		println(key, "=>", value.Name, value.Address, value.User)
	} */

	return nodes
}

// Parse command result
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

// Get a Machine's IP
func GetVmIp(machine Machine) string {
	nodes := loadAllEsxiNodes()
	node := nodes[machine.NodeName]

	buf, err := ioutil.ReadFile(node.IdentityFile)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	config := &ssh.ClientConfig{
		User:            node.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	conn, err := ssh.Dial("tcp", node.Address+":"+node.Port, config)
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

	var sshBuf bytes.Buffer
	session.Stdout = &sshBuf
	remoteCommand := fmt.Sprintf("vim-cmd vmsvc/get.summary %d | grep ipAddress | grep -o [0-9\\.]\\\\+", machine.Id)
	if err := session.Run(remoteCommand); err != nil {
		log.Println(err)
		return ""
	}
	return strings.Replace(sshBuf.String(), "\n", "", -1)
}

// Get All VM Name and VM Id
func GetAllVmIdName() Machines {
	allVm := Machines{}
	for nodeName, nodeInfo := range loadAllEsxiNodes() {
		buf, err := ioutil.ReadFile(nodeInfo.IdentityFile)
		if err != nil {
			log.Fatalln(err)
		}
		key, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			log.Fatalln(err)
		}

		// ssh connect
		nodeAddr := nodeInfo.Address
		nodePort := nodeInfo.Port
		config := &ssh.ClientConfig{
			User:            nodeInfo.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
		}
		b, err := execCommandSsh(nodeAddr, nodePort, config, "vim-cmd vmsvc/getallvms")
		if err != nil {
			log.Println(err.Error())
		}

		// update vm list
		for _, vm := range parseResultAllVms(b) {
			allVm = append(allVm, Machine{
				Id:       vm.Id,
				Name:     vm.Name,
				NodeName: nodeName,
			})
		}
	}
	return allVm
}

func GetVmIpName(ipAddr string) string {
	for _, nodeInfo := range loadAllEsxiNodes() {
		buf, err := ioutil.ReadFile(nodeInfo.IdentityFile)
		if err != nil {
			log.Fatalln(err)
		}
		key, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			log.Fatalln(err)
		}

		// ssh connect
		nodeAddr := nodeInfo.Address
		nodePort := nodeInfo.Port
		config := &ssh.ClientConfig{
			User:            nodeInfo.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
		}
		b, err := execCommandSsh(nodeAddr, nodePort, config, `
			for i in $(vim-cmd vmsvc/getallvms | awk '{print $1}' | grep [0-9]\\+)
			do
				vim-cmd vmsvc/get.summary $i | egrep "\s+(name|ipAddress)" | grep -o '".*"' \
				| sed -e ':a;N;$!ba;s/\n/ /g;s/"//g' | grep [0-9\.]\\{7,\\} &
			done
		`)
		// println(b.String())
		if err != nil {
			log.Println(err.Error())
		}

		// parse ssh result
		for {
			var st, err = b.ReadString('\n')
			if err != nil {
				break
			}

			slice := strings.Split(st, " ")
			vmIp := slice[0]
			vmName := strings.ReplaceAll(strings.Join(slice[1:], "-"), "\n", "")
			if ipAddr == vmIp {
				return vmName
			}
		}
	}
	return ""
}
