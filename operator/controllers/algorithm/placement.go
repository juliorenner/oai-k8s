package algorithm

import (
	"fmt"

	pkgqueue "github.com/Workiva/go-datastructures/queue"
	"github.com/go-logr/logr"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
	v1 "k8s.io/api/core/v1"
)

const (
	logSplitKey = "split"
	dsg8Key     = "1"
)

type Disaggregation interface {
	Validate(ru *oaiv1beta1.RUPosition) (bool, *position)
	AllocateResources(ru *oaiv1beta1.RUPosition) error
}

type PlacementBFS struct {
	root               string
	topology           *oaiv1beta1.Topology
	disaggregations    map[string]*oaiv1beta1.Disaggregation
	requestedResources *utils.RequestedResources
	nodes              map[string]*utils.Node
	cachePaths         map[string][][]string
	log                logr.Logger
}

type pathsValidation struct {
	paths     [][]string
	ruNode    string
	positions map[int]*position
}

type position struct {
	cuNodeName        string
	duNodeName        string
	path              []string
	disaggregationKey string
}

func NewPlacementBFS(topology *oaiv1beta1.Topology, disaggregations map[string]*oaiv1beta1.Disaggregation,
	k8sNodes *v1.NodeList, requestedResources *utils.RequestedResources, log logr.Logger) *PlacementBFS {
	k8sNodeMap := utils.NodeListToMap(k8sNodes)
	core := ""
	graphNodes := make(map[string]*utils.Node)
	for name, nodes := range topology.Nodes {
		k8sNode := k8sNodeMap[name]
		resources := &utils.Resources{
			Memory:          k8sNode.Status.Capacity.Memory(),
			MemoryAvailable: k8sNode.Status.Allocatable.Memory(),
			CPU:             k8sNode.Status.Capacity.Cpu(),
			CPUAvailable:    k8sNode.Status.Allocatable.Cpu(),
		}

		graphNodes[name] = &utils.Node{NodeName: name, Links: make(map[string]*utils.Link), Resources: resources}
		if nodes.Core {
			core = name
		}
	}

	for linkName, link := range topology.Links {
		srcGraph := graphNodes[link.Source.Node]
		dstGraph := graphNodes[link.Destination.Node]
		v := &utils.Link{
			LinkName:           linkName,
			AvailableBandwidth: link.LinkCapacity,
			Latency:            link.LinkDelay,
		}
		srcGraph.Links[link.Destination.Node] = v
		dstGraph.Links[link.Source.Node] = v
	}

	return &PlacementBFS{root: core, topology: topology, nodes: graphNodes, disaggregations: disaggregations,
		requestedResources: requestedResources, log: log}
}

func (p *PlacementBFS) Place(rus []*oaiv1beta1.RUPosition) (bool, error) {
	if err := p.allocateRUsResources(rus); err != nil {
		return false, fmt.Errorf("validation of RU placement failed: %w", err)
	}

	dsg8 := NewDsg8(p.nodes, p.requestedResources, p.disaggregations[dsg8Key])
	for _, ru := range rus {
		paths := p.findPathsTo(ru.RUNode)

		if possible, splitPos := dsg8.Validate(ru, paths); possible {
			fulfillRU(ru, splitPos)

			if err := dsg8.AllocateResources(ru); err != nil {
				return false, fmt.Errorf("error updating resources: %w", err)
			}
			continue
		} else {
			return false, nil
		}
	}

	return true, nil
}

func fulfillRU(ru *oaiv1beta1.RUPosition, finalPos *position) {
	ru.DUNode = finalPos.duNodeName
	ru.CUNode = finalPos.cuNodeName
	ru.Path = finalPos.path
	ru.Disaggregation = finalPos.disaggregationKey
}

func (p *PlacementBFS) findPathsTo(nodeToFind string) [][]string {
	if val, ok := p.cachePaths[nodeToFind]; ok {
		return val
	}

	visited := utils.NewStringSet()
	queue := pkgqueue.New(int64(len(p.nodes)))
	queue.Put([]string{p.root})

	pathsToNode := make([][]string, 0)

	for !queue.Empty() {
		popResult, _ := queue.Get(1)
		path := popResult[0].([]string)
		currentNode := string(path[len(path)-1])
		nodeToExplore := p.nodes[currentNode]
		visited.Add(currentNode)

		if currentNode == nodeToFind {
			pathsToNode = append(pathsToNode, path)
		}

		for nodeName := range nodeToExplore.Links {
			if !visited.Has(nodeName) {
				newPath := getNewPath(path, nodeName)
				queue.Put(newPath)
			}
		}
	}

	return pathsToNode
}

func (p *PlacementBFS) allocateRUsResources(splitsPlacer []*oaiv1beta1.RUPosition) error {
	for _, ru := range splitsPlacer {
		topologyNode := p.nodes[ru.RUNode]
		if err := topologyNode.AllocateResources(p.requestedResources.Memory, p.requestedResources.CPU); err != nil {
			return fmt.Errorf("error allocating split '%s' in node '%s'. Not enough resources available",
				ru.SplitName, ru.RUNode)
		}
		p.log.Info("Able to allocate RU split", logSplitKey, ru.SplitName)
	}

	p.log.Info("Able to allocate all RUs")
	return nil
}

func getNewPath(currentPath []string, newValue string) []string {
	var dst []string
	dst = append(dst, currentPath...)
	return append(dst, newValue)
}
