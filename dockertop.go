package main

import (
	//"bytes"
	//"fmt"
	"log"
	//"os"
	"./docker"
	//"os/exec"
	"strings"
)

var (
	ContainerList []string
)

func updateContainerList() {
	result := docker.Cmd("docker ps | awk '{print $NF}'")
	tmp := strings.Split(result, "\n")
	for _, name := range tmp {
		if len(name) == 0 {
			continue
		}
		if strings.Index(name, "data") > 0 {
			continue
		}
		if name == "NAMES" {
			continue
		}
		ContainerList = append(ContainerList, name)
	}
	log.Printf("get container size %d, names: %s\n", len(ContainerList), strings.Join(ContainerList, "|"))
}

func main() {
	updateContainerList()
	stats, err := docker.Stats("p-uedgpizr")
	if err == nil {
		log.Println(stats)
	}
	tcpStats := docker.TCPStatus("p-uedgpizr")
	log.Println(tcpStats)
}
