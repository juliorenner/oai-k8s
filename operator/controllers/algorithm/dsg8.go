package algorithm

import (
	"fmt"

	pkgqueue "github.com/Workiva/go-datastructures/queue"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
)

type disaggregation8 struct {
	nodes               map[string]*utils.Node
	requestedResources  *utils.RequestedResources
	networkRequirements *oaiv1beta1.Disaggregation
}

func NewDsg8(nodes map[string]*utils.Node, requestedResources *utils.RequestedResources,
	networkRequirements *oaiv1beta1.Disaggregation) *disaggregation8 {
	return &disaggregation8{
		nodes:               nodes,
		requestedResources:  requestedResources,
		networkRequirements: networkRequirements,
	}
}

func (d *disaggregation8) Validate(ru *oaiv1beta1.RUPosition, paths [][]string) (bool, *position) {
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
		return false, nil
	}

	// validate network resources
	for p, candidate := range validation.positions {
		if isValid, _ := d.validateNetwork(paths[p], candidate.cuNodeName, candidate.duNodeName,
			false); isValid {
			return true, candidate
		}
	}

	return false, nil
}

func (d *disaggregation8) validateNetwork(path []string, cuNode, duNode string, allocateResources bool) (bool,
	error) {
	disaggregationQueue := pkgqueue.New(3)
	disaggregationQueue.Put(d.networkRequirements.Backhaul, d.networkRequirements.Midhaul, d.networkRequirements.Fronthaul)

	placementNodes := utils.NewStringSet(path[0], cuNode, duNode)

	var requirement *oaiv1beta1.NetworkRequirements
	// Check if links have the required resources
	for i, nodeName := range path {
		if placementNodes.Has(nodeName) {
			r, _ := disaggregationQueue.Get(1)
			requirement = r[0].(*oaiv1beta1.NetworkRequirements)
		}

		node := d.nodes[nodeName]
		if i+1 < len(path) {
			nextNodeName := path[i+1]
			link := node.Links[nextNodeName]

			if allocateResources {
				if err := link.AllocateResources(requirement.Bandwidth); err != nil {
					return false, fmt.Errorf("error allocating resources: %w", err)
				}
			} else if !link.HasResources(requirement.Bandwidth, requirement.Latency) {
				return false, nil
			}
		}
	}

	return true, nil
}

func (d *disaggregation8) AllocateResources(ru *oaiv1beta1.RUPosition) error {
	// allocate resources from nodes
	du := d.nodes[ru.DUNode]
	if err := du.AllocateResources(d.requestedResources.Memory, d.requestedResources.CPU); err != nil {
		return fmt.Errorf("error allocating du resources: %w", err)
	}

	cu := d.nodes[ru.CUNode]
	if err := cu.AllocateResources(d.requestedResources.Memory, d.requestedResources.CPU); err != nil {
		return fmt.Errorf("error allocating cu resources: %w", err)
	}

	// allocate bandwidth
	if success, err := d.validateNetwork(ru.Path, ru.CUNode, ru.DUNode, true); err != nil || !success {
		return fmt.Errorf("error allocating network resources: %w", err)
	}

	return nil
}