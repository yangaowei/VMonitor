package disk

import (
	"../models"
	"bytes"
	//"fmt"
)

var (
	// store stats for calculating diffs
	vms map[string]VirtualMachineExtended
)

type DiskCollector struct {
}

func (collector DiskCollector) Name() string {
	return "disk"
}

func (collector DiskCollector) Collect(vm models.VirtualMachine) (interface{}, error) {

	result := make(map[string][]int64)

	vmx := vms[vm.Name()]

	// new current counters
	newStats, err := vmx.lookupStats()
	//fmt.Println(newStats)
	if err != nil {
		return result, err
	}

	// get old counters
	oldStats := vmx.Statistic
	if len(oldStats) < 1 {
		vmx.Statistic = newStats
		vms[vm.Name()] = vmx
		return result, nil
	}

	// sum up disk space (available and used)
	// var spaceTotal = make(map[string]int64)
	// var spaceUsed = make(map[string]int64)

	// var spaceTotal int64 = 0
	// var spaceUsed int64 = 0
	for itemName, newStat := range newStats {

		info := make([]int64, 4, 4)
		info[0] = newStat.GetValueAsInt("Capacity")   //spaceTotal
		info[1] = newStat.GetValueAsInt("Allocation") //spaceUsed
		if oldStat, exists := oldStats[itemName]; exists {
			readBandItem := newStat.DiffPerTimeField(oldStat, "rd_bytes")
			info[2] = readBandItem
			writeBandItem := newStat.DiffPerTimeField(oldStat, "wr_bytes")
			info[3] = writeBandItem
		}
		result[itemName] = info
	}
	// calculate read/write bandwidth
	// var readBand int64 = 0
	// var writeBand int64 = 0
	// for itemName, newStat := range newStats {
	// 	if oldStat, exists := oldStats[itemName]; exists {
	// 		readBandItem := newStat.DiffPerTimeField(oldStat, "rd_bytes")
	// 		readBand = readBand + readBandItem
	// 		writeBandItem := newStat.DiffPerTimeField(oldStat, "wr_bytes")
	// 		writeBand = writeBand + writeBandItem
	// 	}
	// }

	// set newStats as oldStats
	vmx.Statistic = newStats
	vms[vm.Name()] = vmx

	// format results
	// spaceTotalMB := (float64(spaceTotal) / 1024 / 1024)
	// spaceUsedMB := (float64(spaceUsed) / 1024 / 1024)
	// readBandMB := (float64(readBand) / 1024 / 1024)
	// writeBandMB := (float64(writeBand) / 1024 / 1024)

	// var result_string = ""
	// for name, value := range result {
	// 	result_string = fmt.Sprintf("%s,%s\t%s\t%.2fMB/s\t%.2fMB/s", name, value[0], value[1], value[2], value[3])
	// 	//fmt.Println(result_string)
	// 	//fmt.Println(name, value)
	// }
	return result, nil
}

func (collector DiskCollector) CollectDetails(vm models.VirtualMachine) {
	// lookup network interfaces for all virtual machines
	list, err := readItems(vm)
	if err != nil {
		return
	}
	if vmx, exists := vms[vm.Name()]; exists {
		vmx.Items = list
		vms[vm.Name()] = vmx
	} else {
		vms[vm.Name()] = VirtualMachineExtended{vm, list, nil}
	}

}

func DefineFlags() {
	//flag.BoolVar(&CPU_EACH, "cpu-each", CPU_EACH, "CPU each")
}

func PrintHeader(buffer *bytes.Buffer) {
	buffer.WriteString("disk-used\t")
	buffer.WriteString("disk-total\t")
	buffer.WriteString("disk-read\t")
	buffer.WriteString("disk-write\t")
}

func Initialize() {
	vms = make(map[string]VirtualMachineExtended)
	models.RegisterCollector(DiskCollector{})
}
