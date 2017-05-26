package network

import (
	"../../util"
	"../models"
	"bytes"
	//"fmt"
	"log"
	"os/exec"
	"regexp"
)

func statisticEvalFunction(statistics models.Statistic) int64 {
	var sum int64 = 0
	if _, exists := statistics.Values["tx_bytes"]; exists {
		sum = sum + statistics.GetValueAsInt("tx_bytes")
	}
	if _, exists := statistics.Values["tx_bytes"]; exists {
		sum = sum + statistics.GetValueAsInt("tx_bytes")
	}
	return sum
}

func statisticDiffPerTime(statsFirst, statsSecond models.Statistic, itemName string) int64 {
	//var sum int64 = 0
	var first int64 = 0
	var second int64 = 0
	if _, exists := statsFirst.Values[itemName]; exists {
		first = statsFirst.GetValueAsInt(itemName)
	}
	if _, exists := statsSecond.Values[itemName]; exists {
		second = statsSecond.GetValueAsInt(itemName)
	}
	diffValue := second - first
	diffTime := int64(statsSecond.Timestamp.Sub(statsFirst.Timestamp).Seconds())
	if diffTime == 0 {
		return 0
	}
	diffPerTime := int64(diffValue / diffTime)
	return diffPerTime
}

type VirtualMachineExtended struct {
	models.VirtualMachine
	Items []models.MeasurementItem
	// key: name of Item
	Statistic map[string]models.Statistic
}

func (vmx *VirtualMachineExtended) lookupStats() (map[string]models.Statistic, error) {
	stats := make(map[string]models.Statistic)
	// loop through vm's interfaces
	for _, vmif := range vmx.Items {
		ifstat, err := readItemStats(vmx.VirtualMachine, vmif)
		if err != nil {
			continue
		}
		stats[vmif.Name] = ifstat
	}
	return stats, nil
}

func listIf(vmname string) []string {
	var cmd *exec.Cmd
	cmd = exec.Command("virsh", "--connect=qemu:///system", "dumpxml", vmname)
	var domifstat bytes.Buffer
	cmd.Stdout = &domifstat
	err := cmd.Run()
	if err != nil {
		log.Printf("Error while exec virsh %a", err)
		//return "error", err
	}
	//fmt.Println(domifstat.String())
	reg := regexp.MustCompile(`vnet\d+`)
	one := reg.FindAllString(domifstat.String(), -1)

	return one

}

func readItems(vm models.VirtualMachine) ([]models.MeasurementItem, error) {
	/*
		# virsh domiflist xy
		Interface  Type       Source     Model       MAC
		-------------------------------------------------------
		tap07e88f58-5d bridge     qbr07e88f58-5d virtio      fa:16:3e:63:1c:a9
	*/
	list, err := util.VirshXList("domiflist", vm.Name())
	if err != nil {
		var itemList []models.MeasurementItem
		iflist := listIf(vm.Name())
		for _, ifitem := range iflist {
			item := models.MeasurementItem{ifitem}
			itemList = append(itemList, item)
		}
		// item = models.MeasurementItem{"vnet1"}
		// itemList = append(itemList, item)
		// item = models.MeasurementItem{"vnet2"}
		// itemList = append(itemList, item)
		list = itemList
	}
	return list, nil
}

func readItemStats(vm models.VirtualMachine, vmif models.MeasurementItem) (models.Statistic, error) {
	/*
		# virsh domiflist xy
		Interface  Type       Source     Model       MAC
		-------------------------------------------------------
		tap07e88f58-5d bridge     qbr07e88f58-5d virtio      fa:16:3e:63:1c:a9
	*/
	stat, err := util.VirshXDetails("domifstat", vm.Name(), vmif.Name, 1, 2, statisticEvalFunction)
	if err != nil {
		return models.Statistic{}, err
	}
	return stat, nil

}
