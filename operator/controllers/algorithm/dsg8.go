package algorithm

import (
	"errors"
	"fmt"

	pkgqueue "github.com/Workiva/go-datastructures/queue"
	"github.com/go-logr/logr"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
)

type disaggregation8 struct {
	nodes               map[string]*utils.Node
	requestedResources  *utils.RequestedResources
	networkRequirements *oaiv1beta1.Disaggregation
	log                 logr.Logger
}

func NewDsg8(nodes map[string]*utils.Node, requestedResources *utils.RequestedResources,
	networkRequirements *oaiv1beta1.Disaggregation, log logr.Logger) *disaggregation8 {
	return &disaggregation8{
		nodes:               nodes,
		requestedResources:  requestedResources,
		networkRequirements: networkRequirements,
		log:                 log,
	}
}

func (d *disaggregation8) Validate(ru *oaiv1beta1.ChainPosition, paths [][]string) (bool, *position) {
	validation := &pathsValidation{
		paths:     paths,
		ruNode:    ru.RUNode,
		positions: map[int]*position{},
	}

	for i, path := range paths {
		v := &position{
			disaggregationKey: dsg8Key,
		}

		// First get the nearest node from the core to place the CU
		v.cuNodeName = path[1]

		// check if the node has resources to run the CU
		cuNode := d.nodes[v.cuNodeName]
		if resourcesAvailable := cuNode.HasResources(d.requestedResources.Memory,
			d.requestedResources.CPU); !resourcesAvailable {
			continue
		}

		for j := 2; j < len(path)-1; j++ {
			duNode := d.nodes[path[j]]
			if resourcesAvailable := duNode.HasResources(d.requestedResources.Memory,
				d.requestedResources.CPU); !resourcesAvailable {
				continue
			}

			v.duNodeName = path[j]
			break
		}

		v.path = path
		validation.positions[i] = v
	}

	if len(validation.positions) == 0 {
		d.log.Error(errors.New("not enough resources remaining"), "no nodes with available resources")
		return false, nil
	}

	// validate network resources
	for p, candidate := range validation.positions {
		if isValid, _ := d.validateNetwork(paths[p], candidate.cuNodeName, candidate.duNodeName,
			false); isValid {
			d.log.Info("found possible allocation", "path", candidate.path)
			return true, candidate
		}
	}

	d.log.Info("no nodes with network requirements", "ru", ru.SplitName)
	return false, nil
}

func (d *disaggregation8) validateNetwork(path []string, cuNode, duNode string, allocateResources bool) (bool,
	error) {
	disaggregationQueue := pkgqueue.New(3)
	disaggregationQueue.Put(d.networkRequirements.Backhaul, d.networkRequirements.Midhaul, d.networkRequirements.Fronthaul)

	placementNodes := utils.NewStringSet(path[0], cuNode, duNode)

	d.log.Info("validating network for path", "path", path)

	var totalLatency float32
	var requirement *oaiv1beta1.NetworkRequirements
	// Check if links have the required resources
	for i, nodeName := range path {
		if placementNodes.Has(nodeName) {
			r, _ := disaggregationQueue.Get(1)
			requirement = r[0].(*oaiv1beta1.NetworkRequirements)
			totalLatency = 0
		}

		node := d.nodes[nodeName]
		if i+1 < len(path) {
			nextNodeName := path[i+1]
			link := node.Links[nextNodeName]

			totalLatency += link.Latency
			d.log.Info("network info", "link", link.LinkName, "available bandwidth", link.AvailableBandwidth,
				"required bandwidth", requirement.Bandwidth,
				"total latency", totalLatency, "required latency", requirement.Latency, "node", nodeName,
				"nextNodeName", nextNodeName)
			if allocateResources {
				if err := link.AllocateResources(requirement.Bandwidth); err != nil {
					return false, fmt.Errorf("error allocating resources: %w", err)
				}
				d.log.Info("remaining link bandwidth", "link", link.LinkName, "bandwidth", link.AvailableBandwidth)
			} else if !link.HasBandwidth(requirement.Bandwidth) ||
				(requirement.Latency > 0 && totalLatency > requirement.Latency) {
				d.log.Info("error: link without required resources", "node", nodeName, "next node", nextNodeName)
				return false, nil
			}
		}
	}

	return true, nil
}

func (d *disaggregation8) AllocateResources(ru *oaiv1beta1.ChainPosition) error {
	// allocate resources from nodes
	du := d.nodes[ru.DUNode]
	if err := du.AllocateResources(d.requestedResources.Memory, d.requestedResources.CPU, d.log); err != nil {
		return fmt.Errorf("error allocating du resources: %w", err)
	}

	cu := d.nodes[ru.CUNode]
	if err := cu.AllocateResources(d.requestedResources.Memory, d.requestedResources.CPU, d.log); err != nil {
		return fmt.Errorf("error allocating cu resources: %w", err)
	}

	// allocate bandwidth
	if success, err := d.validateNetwork(ru.Path, ru.CUNode, ru.DUNode, true); err != nil || !success {
		return fmt.Errorf("error allocating network resources: %w", err)
	}

	return nil
}
