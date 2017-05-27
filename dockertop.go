package main

import (
	//"bytes"
	"./kvm/models"
	"log"
	//"os"
	"./docker"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

var (
	FREQUENCY     = 10
	ContainerList []string
	ContainerMap  map[string]*Container
)

type Container struct {
	Name string
	stat models.Statistic
}

func (container Container) lookupStatus() (statistic models.Statistic, err error) {
	statistic, err = docker.Stats(container.Name)
	return
}

func init() {
	ContainerMap = make(map[string]*Container)

}

func updateContainerList() {
	result := docker.Cmd("docker ps | awk '{print $NF}'")
	tmp := strings.Split(result, "\n")
	ContainerList = []string{}
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
		if _, ok := ContainerMap[name]; !ok {
			container := &Container{name, models.Statistic{}}
			ContainerMap[name] = container
		}
		ContainerList = append(ContainerList, name)
	}
	log.Printf("get container size %d", len(ContainerList))
}

func statisticDiffPerTime(statsFirst, statsSecond models.Statistic) (result map[string]interface{}) {
	diffTime := int64(statsSecond.Timestamp.Sub(statsFirst.Timestamp).Seconds())
	if diffTime == 0 {
		return
	}
	result = make(map[string]interface{})
	firstValues := statsFirst.Values
	secondValues := statsSecond.Values
	result["cpu"] = secondValues["cpu"]
	memory := make(map[string]int)
	total, _ := strconv.Atoi(secondValues["memory_total"])
	used, _ := strconv.Atoi(secondValues["memory_used"])
	memory["total"] = total
	memory["used"] = used
	result["memory"] = memory
	net := make(map[string]int64)
	new_net_in, _ := strconv.ParseInt(secondValues["net_in"], 10, 64)
	first_net_in, _ := strconv.ParseInt(firstValues["net_in"], 10, 64)
	net["in"] = (new_net_in - first_net_in) / diffTime
	new_net_out, _ := strconv.ParseInt(secondValues["net_out"], 10, 64)
	first_net_out, _ := strconv.ParseInt(firstValues["net_out"], 10, 64)
	net["out"] = (new_net_out - first_net_out) / diffTime
	result["net"] = net

	io := make(map[string]int64)
	new_io_in, _ := strconv.ParseInt(secondValues["block_in"], 10, 64)
	first_io_in, _ := strconv.ParseInt(firstValues["block_in"], 10, 64)
	io["in"] = (new_io_in - first_io_in) / diffTime
	new_io_out, _ := strconv.ParseInt(secondValues["block_out"], 10, 64)
	first_io_out, _ := strconv.ParseInt(firstValues["block_out"], 10, 64)
	io["out"] = (new_io_out - first_io_out) / diffTime
	result["io"] = io

	return
}

func main() {
	for {
		start := time.Now()
		nextRun := start.Add(time.Duration(FREQUENCY) * time.Second)
		result := make(map[string]interface{})
		updateContainerList()
		for _, name := range ContainerList {
			container := ContainerMap[name]
			oldStats := container.stat
			newStats, err := container.lookupStatus()
			if err != nil {
				log.Println(err)
				continue
			}
			//tcpStats := docker.TCPStatus(container)
			log.Println(newStats)
			if len(oldStats.Values) < 1 {
				container.stat = newStats
				continue
			}
			//singleResult := make(map[string]interface{})
			singleResult := statisticDiffPerTime(oldStats, newStats)
			singleResult["tcp"] = docker.TCPStatus(container.Name)
			result[name] = singleResult
		}
		if len(result) > 0 {
			log.Println(result, len(result))
			b, _ := json.Marshal(result)
			//bmt.Println(string(b), "value")
			log.Println(string(b))
		}
		time.Sleep(nextRun.Sub(time.Now()))
	}
}
