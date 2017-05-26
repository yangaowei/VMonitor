package main

import (
	"./kvm/cpu"
	"./kvm/disk"
	"./kvm/memory"
	"./kvm/models"
	"./kvm/network"
	"./kvm/static"
	"./util"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	//"github.com/rthornton128/goncurses"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

var (
	// COMPILE VARS
	VERSION = "1.2"

	// CONFIG VARS
	FREQUENCY = 60
	AVERAGE   = false
	//RUNCOUNT       = -1
	RUNCOUNT       = -1
	SHOWIO         = false
	PRINTTIMESTAMP = false
	//QEMU_BINARY_NAME = "kvm"
	QEMU_BINARY_NAME = "qemu-system-x86_64"
	TAB_COLUMN_WIDTH = 18
	SHOW_VERSION     = false
	DEV_UUID         = false
	DEV_CPU          = true
	DEV_MEMORY       = true
	DEV_NETWORK      = true
	DEV_DISK         = true
	OUT_BATCH        = false

	//运行模式
	DEBUG = false

	virtualmachines map[string]models.VirtualMachine
	ppid2vmname     map[string]string

	//screen *goncurses.Window
)

// fill map "virtualmachines" with current vms
func updateVirtualMachineLists() {
	cmd := exec.Command("pidof", QEMU_BINARY_NAME)
	var pids bytes.Buffer
	cmd.Stdout = &pids
	err := cmd.Run()
	if err != nil {
		if !OUT_BATCH {
			//screen.MovePrint(1, 0, fmt.Sprintf("no virtual machines for %s", QEMU_BINARY_NAME))
			fmt.Sprintf("no virtual machines for %s", QEMU_BINARY_NAME)
		}
		//return
	}
	vmPpids := strings.Split(pids.String(), " ")
	// run through all VMs' parent process ids (ppid)
	for _, ppid := range vmPpids {
		ppid = strings.Replace(ppid, "\n", "", -1)
		if ppid == "" {
			continue
		}
		// if vm is known
		if vmName, ok := ppid2vmname[ppid]; ok {
			vm := virtualmachines[vmName]
			// collect VM details in non-force-mode
			vm.CollectDetails(false)
		} else {
			// if vm is NOT known, create one
			vm := models.CreateVM(ppid)
			// collect VM details in force-mode
			vm.CollectDetails(true)
			// store VM in lists for later lookup
			virtualmachines[vm.Name()] = vm
			ppid2vmname[ppid] = vm.Name()
		}
	}
}

// Get the utilisation measurements and print them
func measureVirtualMachines(runs int) {

	var printerBuffer bytes.Buffer

	// format output in tabs
	var testBuffer bytes.Buffer
	testBufferWriter := bufio.NewWriter(&testBuffer)
	w := new(tabwriter.Writer)
	if OUT_BATCH {
		w.Init(testBufferWriter, 0, 8, 1, '\t', 0)
	} else {
		w.Init(testBufferWriter, 0, 8, 1, ' ', 0)
	}

	// get header row (always except in batch mode: only once)
	if !OUT_BATCH || (OUT_BATCH && runs < 0) {
		printHeader(&printerBuffer)
		fmt.Fprintln(w, printerBuffer.String())
	}

	// walk through vms and measure
	n := 0
	for i, vm := range virtualmachines {
		printerBuffer.Reset()
		printerBuffer.WriteString(vm.Name())
		printerBuffer.WriteString("\t")
		// call each registered collector
		//fmt.Println(i, vm)
		for _, collector := range models.GetCollectors() {
			// TODO parallelise the collector.Collect block
			_, err := collector.Collect(vm)
			if err != nil {
				delete(ppid2vmname, vm.Ppid())
				delete(virtualmachines, i)
				continue
			}
			//printerBuffer.WriteString(value)
			printerBuffer.WriteString("\t")
		}
		if runs >= 0 {
			//screen.MovePrint(n+1, 0, buffer.String())
			fmt.Fprintln(w, printerBuffer.String())
		} else {
			if !OUT_BATCH {
				fmt.Fprintln(w, "measuring...")
			}
		}
		n++
	}

	// print
	w.Flush()
	testBufferWriter.Flush()
	fmt.Print(testBuffer.String())
}

func defineFlags() {
	// general flags
	flag.IntVar(&FREQUENCY, "s", FREQUENCY, "sleep n seconds between runs. default 1s")
	flag.IntVar(&RUNCOUNT, "r", RUNCOUNT, "runs x times then terminates. default -1 (runs forever)")
	flag.StringVar(&QEMU_BINARY_NAME, "qemu-binary-name", QEMU_BINARY_NAME, "binary name of qemu driver. default qemu-kvm")
	flag.BoolVar(&SHOW_VERSION, "version", SHOW_VERSION, "show version")

	// selection of devices
	flag.BoolVar(&DEV_UUID, "uuid", DEV_UUID, "show uuid  (default: false)")
	flag.BoolVar(&DEV_CPU, "cpu", DEV_CPU, "show cpu (default: false)")
	flag.BoolVar(&DEV_MEMORY, "memory", DEV_MEMORY, "show memory (default: true)")
	flag.BoolVar(&DEV_NETWORK, "network", DEV_NETWORK, "show network (default: false)")
	flag.BoolVar(&DEV_DISK, "disk", DEV_DISK, "show disk  (default: false)")

	// control output
	flag.BoolVar(&OUT_BATCH, "batch", OUT_BATCH, "use simple output e.g. for scripts (default: false)")

	//debug
	flag.BoolVar(&DEBUG, "debug", DEBUG, "debug (default: true)")

	// define flags for each collector
	static.DefineFlags()
	cpu.DefineFlags()
	memory.DefineFlags()
	network.DefineFlags()
	disk.DefineFlags()

	flag.Parse()
}

func printHeader(buffer *bytes.Buffer) {
	buffer.WriteString("vmname\t")
	if DEV_UUID {
		static.PrintHeader(buffer)
	}
	if DEV_CPU {
		cpu.PrintHeader(buffer)
	}
	if DEV_MEMORY {
		memory.PrintHeader(buffer)
	}
	if DEV_NETWORK {
		network.PrintHeader(buffer)
	}
	if DEV_DISK {
		disk.PrintHeader(buffer)
	}
}

func initialize() {
	// catche SIGINT signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		shutdown()
	}()

	if DEV_UUID {
		static.Initialize()
	}
	if DEV_CPU {
		cpu.Initialize()
	}
	if DEV_MEMORY {
		memory.Initialize()
	}
	if DEV_NETWORK {
		network.Initialize()
	}
	if DEV_DISK {
		disk.Initialize()
	}

	// create arrays for storing measurements
	virtualmachines = make(map[string]models.VirtualMachine)
	ppid2vmname = make(map[string]string)

	// first initial querying of measurements
	updateVirtualMachineLists()
}

func debug_main() {

	defineFlags()
	if SHOW_VERSION {
		fmt.Println("kvmtop version " + VERSION)
		return
	}

	initialize()
	//count := 1
	for n := -1; RUNCOUNT == -1 || n < RUNCOUNT; n++ {
		start := time.Now()
		measureVirtualMachines(n)
		nextRun := start.Add(time.Duration(FREQUENCY) * time.Second)
		updateVirtualMachineLists()
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

func getMachinesInfo(runs int) string {
	//result := make(map[string]interface{})
	//err := nil
	//fmt.Println(virtualmachines)
	vm_info := make(map[string]interface{})
	for i, vm := range virtualmachines {
		var value interface{}
		var err error
		item_info := make(map[string]interface{})
		for _, collector := range models.GetCollectors() {
			// TODO parallelise the collector.Collect block
			value, err = collector.Collect(vm)
			if err != nil {
				delete(ppid2vmname, vm.Ppid())
				delete(virtualmachines, i)
				continue
			}
			item_info[collector.Name()] = value
		}
		vm_info[i] = item_info
	}
	if runs >= 0 {
		//screen.MovePrint(n+1, 0, buffer.String())
		b, _ := json.Marshal(vm_info)
		//bmt.Println(string(b), "value")
		return string(b)
	}
	return ""
}

func main() {
	if DEBUG {
		debug_main()
	} else {
		fmt.Println("starting")
		defineFlags()
		initialize()
		for n := -1; RUNCOUNT == -1 || n < RUNCOUNT; n++ {
			start := time.Now()
			nextRun := start.Add(time.Duration(FREQUENCY) * time.Second)
			value := getMachinesInfo(n)
			if value != "" {
				log.Printf("collect %d machines info", len(value))
				//发送数据
				log.Println(value)
				util.SendData(time.Now().Unix(), value)
			}
			updateVirtualMachineLists()
			time.Sleep(nextRun.Sub(time.Now()))
		}
	}
}

func shutdown() {
	os.Exit(1)
}
