package algorithm

import (
	"encoding/json"
	"fmt"
	"testing"

	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var topologyJSON = "{\r\n    \"nodes\": {\r\n        \"node1\": {\r\n            \"interfaces\": [\"eth0\", " +
	"\"eth1\", \"eth2\", \"eth3\"],\r\n            \"hops\": 1\r\n        },\r\n        \"node2\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\", \"eth2\", \"eth3\", \"eth4\"],\r\n            \"hops\": 1\r\n        },\r\n        \"node3\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\", \"eth2\", \"eth3\", \"eth4\"],\r\n            \"hops\": 2\r\n        },\r\n        \"node4\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\", \"eth2\", \"eth3\"],\r\n            \"hops\": 2\r\n        },\r\n        \"node5\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 2\r\n        },\r\n        \"node6\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 3\r\n        },\r\n        \"node7\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 3\r\n        },\r\n        \"node8\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 3\r\n        },\r\n        \"node9\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 4\r\n        },\r\n        \"node10\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\", \"eth2\"],\r\n            \"hops\": 3\r\n        },\r\n        \"node11\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 3\r\n        },\r\n        \"node12\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 4\r\n        },\r\n        \"node13\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"hops\": 5\r\n        },\r\n        \"node14\": {\r\n            \"interfaces\": [\"eth0\", \"eth1\"],\r\n            \"core\": true\r\n        }\r\n    },\r\n    \"links\": {\r\n        \"node6--node3\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node6\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node3\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node6--node4\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node6\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node4\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node7--node3\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node7\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node3\",\r\n                \"interface\": \"eth3\"\r\n            }\r\n        },\r\n        \"node7--node5\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node7\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node5\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node8--node3\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node8\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node3\",\r\n                \"interface\": \"eth4\"\r\n            }\r\n        },\r\n        \"node9--node8\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node9\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node8\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        },\r\n        \"node10--node9\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node10\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node9\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        },\r\n        \"node10--node4\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node10\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node4\",\r\n                \"interface\": \"eth3\"\r\n            }\r\n        },\r\n        \"node11--node4\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node11\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node4\",\r\n                \"interface\": \"eth4\"\r\n            }\r\n        },\r\n        \"node11--node5\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node11\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node5\",\r\n                \"interface\": \"eth3\"\r\n            }\r\n        },\r\n        \"node12--node11\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 0.25,\r\n            \"source\": {\r\n                \"node\": \"node12\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node11\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node13--node12\": {\r\n            \"linkCapacity\": 300,\r\n            \"LinkDelay\": 1,\r\n            \"source\": {\r\n                \"node\": \"node13\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node12\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        },\r\n        \"node3--node1\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 3,\r\n            \"source\": {\r\n                \"node\": \"node3\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node1\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        },\r\n        \"node3--node2\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 4,\r\n            \"source\": {\r\n                \"node\": \"node3\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node2\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        },\r\n        \"node4--node1\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 5,\r\n            \"source\": {\r\n                \"node\": \"node4\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node1\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node4--node2\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 6,\r\n            \"source\": {\r\n                \"node\": \"node4\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node2\",\r\n                \"interface\": \"eth2\"\r\n            }\r\n        },\r\n        \"node5--node1\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 3,\r\n            \"source\": {\r\n                \"node\": \"node5\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node1\",\r\n                \"interface\": \"eth3\"\r\n            }\r\n        },\r\n        \"node5--node2\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 4,\r\n            \"source\": {\r\n                \"node\": \"node5\",\r\n                \"interface\": \"eth1\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node2\",\r\n                \"interface\": \"eth3\"\r\n            }\r\n        },\r\n        \"node1--node14\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 2,\r\n            \"source\": {\r\n                \"node\": \"node1\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node14\",\r\n                \"interface\": \"eth0\"\r\n            }\r\n        },\r\n        \"node2--node14\": {\r\n            \"linkCapacity\": 1200,\r\n            \"LinkDelay\": 3,\r\n            \"source\": {\r\n                \"node\": \"node2\",\r\n                \"interface\": \"eth0\"\r\n            },\r\n            \"destination\": {\r\n                \"node\": \"node14\",\r\n                \"interface\": \"eth1\"\r\n            }\r\n        }\r\n    }\r\n}"

var disaggregationJSON = "{\r\n    \"1\": {\r\n        \"protocolStack\": {\r\n            \"cu\": [\"RRC\", " +
	"\"PDCP\"],\r\n            \"du\": [\"RLCH\", \"RLCL\", \"MACH\", \"MACL\"],\r\n            \"ru\": [\"PHYH\", \"PHYL\", \"RF\"]\r\n        },\r\n        \"splitOptions\": {\r\n            \"cu-du\": \"O2\",\r\n            \"du-ru\": \"O6\"\r\n        },\r\n        \"backhaul\": {\r\n            \"bandwidth\": 151\r\n        },\r\n        \"midhaul\": {\r\n            \"latency\": 30,\r\n            \"bandwidth\": 151\r\n        },\r\n        \"fronthaul\": {\r\n            \"latency\": 2,\r\n            \"bandwidth\": 152\r\n        },\r\n        \"crosshaul\": {\r\n            \"latency\": 30\r\n        }\r\n    },\r\n    \"2\": {\r\n        \"protocolStack\": {\r\n            \"cu\": [\"RRC\", \"PDCP\"],\r\n            \"du\": [],\r\n            \"ru\": [\"RLCH\", \"RLCL\", \"MACH\", \"MACL\", \"PHYH\", \"PHYL\", \"RF\"]\r\n        },\r\n        \"backhaul\": {\r\n            \"bandwidth\": 151\r\n        },\r\n        \"midhaul\": {},\r\n        \"fronthaul\": {\r\n            \"bandwidth\": 151\r\n        },\r\n        \"crosshaul\": {\r\n            \"latency\": 30\r\n        }\r\n    },\r\n    \"3\": {\r\n        \"protocolStack\": {\r\n            \"cu\": [\"RRC\", \"PDCP\", \"RLCH\", \"RLCL\", \"MACH\", \"MACL\"],\r\n            \"du\": [],\r\n            \"ru\": [\"PHYH\", \"PHYL\", \"RF\"]\r\n        },\r\n        \"backhaul\": {\r\n            \"bandwidth\": 151\r\n        },\r\n        \"midhaul\": {},\r\n        \"fronthaul\": {\r\n            \"latency\": 2,\r\n            \"bandwidth\": 152\r\n        },\r\n        \"crosshaul\": {\r\n            \"latency\": 30\r\n        }\r\n    },\r\n    \"4\": {\r\n        \"protocolStack\": {\r\n            \"cu\": [\"RRC\", \"PDCP\", \"RLCH\", \"RLCL\", \"MACH\", \"MACL\", \"PHYH\", \"PHYL\", \"RF\"],\r\n            \"du\": [],\r\n            \"ru\": []\r\n        },\r\n        \"backhaul\": {},\r\n        \"midhaul\": {},\r\n        \"fronthaul\": {},\r\n        \"crosshaul\": {\r\n            \"latency\": 30\r\n        }\r\n    }\r\n}"

func TestQuantity(t *testing.T) {
	memory := utils.NewQuantity("500Mi")
	cpu := utils.NewQuantity("500m")

	memoryNode := utils.NewQuantity("16397940Ki")
	cpuNode := utils.NewQuantity("3800m")

	fmt.Println(memoryNode.ScaledValue(resource.Mega))

	memoryNode.Sub(*memory)
	cpuNode.Sub(*cpu)

	fmt.Println(memoryNode.Value())
	fmt.Println(cpuNode.Value())
	fmt.Println(memoryNode.ScaledValue(resource.Mega))
	fmt.Println(cpuNode)
}

func TestPlacementAlgorithm(t *testing.T) {
	disaggregation := map[string]*oaiv1beta1.Disaggregation{}

	if err := json.Unmarshal([]byte(disaggregationJSON), &disaggregation); err != nil {
		t.Fatalf("error unmarshaling disaggregation: %s", err)
	}

	log := zap.New(zap.UseDevMode(true))
	k8sNode := generateNodeList()
	requestedResources := &utils.RequestedResources{
		Memory: *utils.NewQuantity("512Mi"),
		CPU:    *utils.NewQuantity("500m"),
	}

	testCases := []struct {
		name                 string
		rus                  []*oaiv1beta1.ChainPosition
		isValid              bool
		isErrorExpected      bool
		numberOfAllocatedRUs int
	}{
		{
			"no enough link", generateRUs("node6", "node6", "node6"), true, false, 2,
		},
		{
			"enough resources", generateRUs("node6", "node6"), true, false, 2,
		},
		{
			"enough resources: node13", generateRUs("node13"), true, false, 1,
		},
		{
			"not enough resources: node13", generateRUs("node13", "node13"), true, false, 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterTestingT(t)

			topology := &oaiv1beta1.Topology{}
			if err := json.Unmarshal([]byte(topologyJSON), topology); err != nil {
				t.Fatalf("error unmarshaling topology: %s", err)
			}
			topologyGraph := NewPlacementBFS(topology, disaggregation, k8sNode, requestedResources, log)

			valid, err := topologyGraph.Place(tc.rus)
			Expect(valid).To(Equal(tc.isValid))
			Expect(err != nil).To(Equal(tc.isErrorExpected))

			count := 0
			for _, ru := range tc.rus {
				if count < tc.numberOfAllocatedRUs {
					Expect(ru.DUNode).To(Not(BeEmpty()))
					Expect(ru.CUNode).To(Not(BeEmpty()))
				} else {
					Expect(ru.DUNode).To(BeEmpty())
					Expect(ru.CUNode).To(BeEmpty())
				}
				count += 1
			}
		})
	}
}

func generateRUs(nodes ...string) []*oaiv1beta1.ChainPosition {
	var rus []*oaiv1beta1.ChainPosition
	for i, node := range nodes {
		rus = append(rus, &oaiv1beta1.ChainPosition{
			SplitName: fmt.Sprintf("split%d", i),
			RUNode:    node,
		})
	}

	return rus
}

func generateNodeList() *v1.NodeList {
	nodeList := &v1.NodeList{}
	for i := 1; i < 20; i++ {
		node := v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("node%d", i),
			},
			Status: v1.NodeStatus{
				Capacity: v1.ResourceList{
					v1.ResourceCPU:    *utils.NewQuantity("5000m"),
					v1.ResourceMemory: *utils.NewQuantity("8192Mi"),
				},
				Allocatable: v1.ResourceList{
					v1.ResourceCPU:    *utils.NewQuantity("4000m"),
					v1.ResourceMemory: *utils.NewQuantity("6144Mi"),
				},
			},
		}

		nodeList.Items = append(nodeList.Items, node)
	}

	return nodeList
}
