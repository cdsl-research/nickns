package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
)

type HostConfig struct {
	Name string
	Address string
	User string
	IdentityFile string `toml:"identity_file"`
	Password string
}

var HostsConfig map[string]HostConfig

func main() {
	content, err := ioutil.ReadFile("./sample.toml")
	if err != nil {
		panic(err)
	}

	if _, err := toml.Decode(string(content), &HostsConfig); err != nil {
		log.Fatal(err)
	}

	for key,value := range HostsConfig {
		println("key=", key, "value=", value.Name, value.Address, value.User, value.Password, value.IdentityFile)
	}
}