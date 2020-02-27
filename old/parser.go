package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Machine struct {
	Id         int
	Name       string
	File       string
	GuestOs    string
	Version    string
	Annotation string
}

type Machines []Machine

func parser() {
	File Open
	file, err := os.Open(`./raw`)
	if err != nil {
		log.Fatal("Could not open file: ", err.Error())
	}
	defer file.Close()

	// Regex
	r := regexp.MustCompile(`^\d.+`)

	// Parse
	var vms Machines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line_str := scanner.Text()

		// Match Regex
		if r.MatchString(line_str) {
			slice := strings.Split(line_str, "    ")
			slice0, err := strconv.Atoi(slice[0])
			if err != nil {
				slice0 = -1
			}
			vm := Machine{
				Id:   slice0,
				Name: strings.TrimSpace(slice[1]),
			}
			vms = append(vms, vm)
			// fmt.Println(vm.Id, vm.Name)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("File Scanner Error: ", err.Error())
	}

	// Print
	// fmt.Println(vms)
}

func main() {
  parser()
}
