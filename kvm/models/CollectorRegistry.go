package models

type Collector interface {
	Collect(vm VirtualMachine) (interface{}, error)
	CollectDetails(vm VirtualMachine)
	Name() string
}

var (
	collectors []Collector
)

func RegisterCollector(collector Collector) {
	collectors = append(collectors, collector)
}
func GetCollectors() []Collector {
	return collectors
}
