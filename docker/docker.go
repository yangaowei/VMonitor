package docker

import (
	"../kvm/models"
	"bytes"
	"log"
	"os/exec"
	//"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Cmd(cmds string) (result string) {
	log.Println("run cmd:", cmds)
	var cmd *exec.Cmd
	cmd = exec.Command("/bin/sh", "-c", cmds)
	var domifstat bytes.Buffer
	cmd.Stdout = &domifstat
	err := cmd.Run()
	if err != nil {
		log.Printf("Error while exec cmd %a", err)
		return ""
	}
	result = domifstat.String()
	return
}

func RxOf(pattern string, content string, index int) (rcontent string) {
	re, _ := regexp.Compile(pattern)
	submatch := re.FindStringSubmatch(content)
	for i, v := range submatch {
		if i == index {
			rcontent = v
			break
		}
	}
	return
}

func R1(pattern string, content string) (rcontent string) {
	return RxOf(pattern, content, 1)
}

func evalFunc(models.Statistic) int64 {
	return 100
}

func parseBase(s string) (num, unit string) {
	num = R1("([0-9]+)", s)
	unit = R1("([a-zA-Z]+)", s)
	return
}

//网络，读写单位统一B
func parseNetB(s string) string {
	num, unit := parseBase(s)
	if unit == "B" {
		return num
	}
	tmp, _ := strconv.Atoi(num)
	if unit == "kB" {
		tmp = tmp * 1024
	}
	if unit == "MB" {
		tmp = tmp * 1024 * 1024
	}
	num = strconv.Itoa(tmp)
	return num
}

//内存单位统一为MiB
func parseMemory(memory string) string {
	num, unit := parseBase(memory)
	if unit == "GiB" {
		tmp, _ := strconv.Atoi(num)
		tmp = tmp * 1024
		num = strconv.Itoa(tmp)
	}
	return num
}

func TCPStatus(container string) (result map[string]int) {
	result = make(map[string]int)
	var buffer bytes.Buffer
	buffer.WriteString("docker exec ")
	buffer.WriteString(container)
	buffer.WriteString(" netstat -na|awk '/^tcp/ {++S[$NF]}END{for(a in S) print a,S[a]}'")
	tmp := Cmd(buffer.String())
	for _, value := range strings.Split(tmp, "\n") {
		if len(value) == 0 {
			continue
		}
		kv := strings.Split(value, " ")
		num, _ := strconv.Atoi(kv[1])
		result[kv[0]] = num
	}
	return
}

func Stats(container string) (models.Statistic, error) {
	var buffer bytes.Buffer
	buffer.WriteString("docker stats --no-stream=true ")
	buffer.WriteString(container)
	buffer.WriteString("|grep ")
	buffer.WriteString(container)
	result := Cmd(buffer.String())
	if result == "" {
		log.Println(result)
	} else {
		reg := regexp.MustCompile(" +")
		tmp := reg.Split(result, 100)
		// for index, value := range tmp {
		// 	log.Println(index, value)
		// }
		stats := make(map[string]string)
		stats["cpu"] = tmp[1]
		stats["memory_used"] = parseMemory(tmp[2] + tmp[3])
		stats["memory_total"] = parseMemory(tmp[5] + tmp[6])
		stats["net_in"] = parseNetB(tmp[8] + tmp[9])
		stats["net_out"] = parseNetB(tmp[11] + tmp[12])
		stats["block_in"] = parseNetB(tmp[13] + tmp[14])
		stats["block_out"] = parseNetB(tmp[16] + tmp[17])
		stat := models.Statistic{time.Now(), stats, evalFunc}
		return stat, nil
	}
	return models.Statistic{}, nil
}
