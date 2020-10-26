package utils

import (
	"fmt"

	"github.com/go-logr/logr"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type StringSet map[string]struct{}

type Placement interface {
	// Place should fulfill each RUPosition according to the implement algorithm results
	Place([]*oaiv1beta1.RUPosition) (bool, error)
}

var Empty struct{}

func NewStringSet(values ...string) StringSet {
	stringSet := make(StringSet)
	for _, v := range values {
		stringSet[v] = Empty
	}
	return stringSet
}

// Add adds new values to the set.
func (s *StringSet) Add(items ...string) {
	for _, item := range items {
		(*s)[item] = Empty
	}
}

// Has returns true if item is in the Set
func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}

func NewMemoryQuantity(value int64) *resource.Quantity {
	return resource.NewQuantity(value, resource.BinarySI)
}

func NewCPUQuantity(value int64) *resource.Quantity {
	return resource.NewQuantity(value, resource.DecimalSI)
}

type RequestedResources struct {
	Memory resource.Quantity
	CPU    resource.Quantity
}

type Node struct {
	NodeName string
	// key is the Node name
	Links     map[string]*Link
	Resources *Resources
}

func (node *Node) HasResources(memory, cpu resource.Quantity) bool {
	return node.Resources.MemoryAvailable.Value() > memory.Value() && node.Resources.CPUAvailable.Value() > cpu.
		Value()
}

func (node *Node) AllocateResources(memory, cpu resource.Quantity, log logr.Logger) error {
	log.Info("memory", "mem", node.Resources.MemoryAvailable.Value())
	log.Info("cpu", "cpu_1", node.Resources.CPUAvailable.Value())
	node.Resources.MemoryAvailable.Sub(memory)
	node.Resources.CPUAvailable.Sub(cpu)

	if node.Resources.CPUAvailable.Value() < 0 ||
		node.Resources.MemoryAvailable.Value() < 0 {
		return fmt.Errorf("error allocating resources. CPU: %d, Memory: %d", node.Resources.CPUAvailable.Value(),
			node.Resources.MemoryAvailable.Value())
	}

	return nil
}

type Resources struct {
	Memory          *resource.Quantity
	MemoryAvailable *resource.Quantity
	CPU             *resource.Quantity
	CPUAvailable    *resource.Quantity
}

type Link struct {
	LinkName           string
	AvailableBandwidth float32
	Latency            float32
}

func (v *Link) HasResources(requiredBandwidth, requiredLatency float32) bool {
	if requiredLatency == 0 {
		return v.AvailableBandwidth >= requiredBandwidth
	}
	return v.AvailableBandwidth >= requiredBandwidth && v.Latency <= requiredLatency
}

func (v *Link) AllocateResources(requiredBandwidth float32) error {
	v.AvailableBandwidth -= requiredBandwidth
	if v.AvailableBandwidth < 0 {
		return fmt.Errorf("link '%s' does not have enough bandwidht", v.LinkName)
	}
	return nil
}
