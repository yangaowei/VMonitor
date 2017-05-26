package network

import (
	"../models"
	"bytes"
	//"fmt"
)

var (
	// store stats for calculating diffs
	vms map[string]VirtualMachineExtended
)

type NetworkCollector struct {
}

func (collector NetworkCollector) Name() string {
	return "Network"
}

func (collector NetworkCollector) Collect(vm models.VirtualMachine) (interface{}, error) {
	vmx := vms[vm.Name()]

	// new current counters
	newStats, err := vmx.lookupStats()
	if err != nil {
		return "", err
	}

	// get old counters
	oldStats := vmx.Statistic
	//fmt.Println(newStats)
	if len(oldStats) < 1 {
		vmx.Statistic = newStats
		vms[vm.Name()] = vmx
		return "-", nil
	}

	// calculate diff between new and old counters
	//var utilSum map[string]int64
	result := make(map[string][]int64)
	for itemName, newStat := range newStats {
		if oldStat, exists := oldStats[itemName]; exists {
			tmp := make([]int64, 2, 2)
			tmp[0] = statisticDiffPerTime(oldStat, newStat, "tx_bytes")
			tmp[1] = statisticDiffPerTime(oldStat, newStat, "rx_bytes")
			result[itemName] = tmp
		}
	}

	// set newStats as oldStats
	vmx.Statistic = newStats
	vms[vm.Name()] = vmx

	//utilMB := (float64(utilSum) / 1024 / 1024)
	//result := fmt.Sprintf("%s-%s", timeTamps, utilSum)
	return result, nil

}

func (collector NetworkCollector) CollectDetails(vm models.VirtualMachine) {
	// lookup network interfaces for all virtual machines
	iflist, err := readItems(vm)
	//fmt.Println(iflist, vm, "iflist")
	if err != nil {
		return
	}
	if vmx, exists := vms[vm.Name()]; exists {
		vmx.Items = iflist
		vms[vm.Name()] = vmx
	} else {
		vms[vm.Name()] = VirtualMachineExtended{vm, iflist, nil}
	}

}

func DefineFlags() {
	//flag.BoolVar(&CPU_EACH, "cpu-each", CPU_EACH, "CPU each")
}

func PrintHeader(buffer *bytes.Buffer) {
	buffer.WriteString("network\t")
}

func Initialize() {
	vms = make(map[string]VirtualMachineExtended)
	models.RegisterCollector(NetworkCollector{})
}
