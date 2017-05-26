package cpu

import (
	"fmt"
	//"flag"
	"../models"
	"bytes"
)

var (
	// store stats for calculating diffs
	cpustats map[string]VmCpuSchedStat
	CPU_EACH = false
)

type CpuCollector struct {
}

func (collector CpuCollector) Name() string {
	return "CPU"
}

func (collector CpuCollector) Collect(vm models.VirtualMachine) (interface{}, error) {
	// new current CPU counters
	result := make([]interface{}, 4, 4)
	newSchedStat, err := lookupStats(vm)
	if err != nil {
		return result, err
	}

	// get old CPU countersd
	if oldSchedStat, exists := cpustats[vm.Ppid()]; exists {
		// set new stats as old stats for next run
		cpustats[vm.Ppid()] = newSchedStat
		// calculate diff between new and old counters
		cpu_utilisation := newSchedStat.calculateDiff(oldSchedStat)
		vCores := len(vm.VCpuTasks())
		// result := fmt.Sprintf("%d\t%.0f%%\t%.0f%%\t%.0f%%\t%.0f%%",
		// 	vCores,
		// 	(cpu_utilisation.Avg.Inside * 100),
		// 	(cpu_utilisation.Avg.Outside * 100),
		// 	(cpu_utilisation.Avg.Steal * 100),
		// 	(cpu_utilisation.Avg_other * 100))

		result[0] = vCores
		result[1] = cpu_utilisation.Avg.Inside * 100
		result[2] = cpu_utilisation.Avg.Outside * 100
		result[3] = cpu_utilisation.Avg.Steal * 100
	} else {
		// no measurement yet
		// set new stats as old stats for next run
		cpustats[vm.Ppid()] = newSchedStat
	}
	return result, nil
}

func (collector CpuCollector) CollectDetails(vm models.VirtualMachine) {
	// TODO lookup vCores here! not in VirtualMachine
}

func DefineFlags() {
	//flag.BoolVar(&CPU_EACH, "cpu-each", CPU_EACH, "CPU each")
}

func PrintHeader(buffer *bytes.Buffer) {
	buffer.WriteString("CpuCS\t")
	buffer.WriteString("CpuVM\t")
	buffer.WriteString("CpuPM\t")
	buffer.WriteString("CpuST\t")
	buffer.WriteString("CpuIO\t")
}

func Initialize() {
	if CPU_EACH {
		fmt.Println("PRINT EACH CPU")
	}

	cpustats = make(map[string]VmCpuSchedStat)
	models.RegisterCollector(CpuCollector{})
}
