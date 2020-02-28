package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
)

func main() {
	buf, err := ioutil.ReadFile("./id_rsa")
	if err != nil {
		panic(err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}

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

	conn, err := ssh.Dial("tcp", ip+":"+port, config)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Println(err)
	}
	defer session.Close()

	//Check whoami
	var b bytes.Buffer
	session.Stdout = &b
	remote_command := "date"
	if err := session.Run(remote_command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	log.Println(remote_command + ":" + b.String())
}
