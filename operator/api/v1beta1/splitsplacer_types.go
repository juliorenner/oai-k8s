/*
Copyright 2020 Julio Renner.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PlacerStateFinished = "Finished"
	PlacerStateError    = "Error"

	RRC  DisaggregationProtocolStack = "RRC"
	PDCP DisaggregationProtocolStack = "PDCP"
	RLCH DisaggregationProtocolStack = "RLCH"
	RLCL DisaggregationProtocolStack = "RLCL"
	MACH DisaggregationProtocolStack = "MACH"
	MACL DisaggregationProtocolStack = "MACL"
	PHYH DisaggregationProtocolStack = "PHYH"
	PHYL DisaggregationProtocolStack = "PHYL"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SplitsPlacerState string
type DisaggregationProtocolStack string

// SplitsPlacerSpec defines the desired state of SplitsPlacer
type SplitsPlacerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// RUs
	// +kubebuilder:validation:Required
	RUs []*RUPosition `json:"rus,omitempty"`
	// CoreIP to where the splits created will point to.
	// +kubebuilder:validation:Required
	CoreIP string `json:"coreIP,omitempty"`
	// Topology refers to the config map name where the topology is described
	TopologyConfig string `json:"topologyConfig,omitempty"`
	// Retrigger placement
	Retrigger bool `json:"retrigger,omitempty"`
}

// RUPosition defines the position and the name of the RU from one service chain. Based on this definition a Split
// will be created.
type RUPosition struct {
	SplitName string `json:"splitName,omitempty"`
	RUNode    string `json:"ruNode,omitempty"`
	// CUNode will be fulfilled by the split placer algorithm
	CUNode string `json:"cuNode,omitempty"`
	// DUNode will be fulfilled by the split placer algorithm
	DUNode string `json:"duNode,omitempty"`
	// Path will be fulfilled by the split placer algorithm
	Path []string `json:"path,omitempty"`
	// Disaggregation will be fulfilled by the split placer algorithm
	Disaggregation string `json:"disaggregation,omitempty"`
}

// SplitsPlacerStatus defines the observed state of SplitsPlacer
type SplitsPlacerStatus struct {
	State              SplitsPlacerState  `json:"state,omitempty"`
	RemainingBandwidth map[string]float32 `json:"remainingBandwidth,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.state`
// SplitsPlacer is the Schema for the splitsplacers API
type SplitsPlacer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SplitsPlacerSpec   `json:"spec,omitempty"`
	Status SplitsPlacerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SplitsPlacerList contains a list of SplitsPlacer
type SplitsPlacerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SplitsPlacer `json:"items"`
}

// Topology is the Schema for the nodes topology where the splits will be placed
type Topology struct {
	Nodes map[string]*Node `json:"nodes,omitempty"`
	Links map[string]*Link `json:"links,omitempty"`
}

type Node struct {
	Interfaces []string `json:"interfaces,omitempty"`
	Core       bool     `json:"core,omitempty"`
	Hops       int      `json:"hops,omitempty"`
}

type Link struct {
	LinkCapacity float32    `json:"linkCapacity,omitempty"`
	LinkDelay    float32    `json:"linkDelay,omitempty"`
	Source       Connection `json:"source,omitempty"`
	Destination  Connection `json:"destination,omitempty"`
}

type Connection struct {
	Node      string `json:"node,omitempty"`
	Interface string `json:"interface,omitempty"`
}

type Disaggregation struct {
	ProtocolStack ProtocolStack        `json:"protocolStack,omitempty"`
	Backhaul      *NetworkRequirements `json:"backhaul,omitempty"`
	Midhaul       *NetworkRequirements `json:"midhaul,omitempty"`
	Fronthaul     *NetworkRequirements `json:"fronthaul,omitempty"`
}

type ProtocolStack struct {
	CU []DisaggregationProtocolStack `json:"cu,omitempty"`
	DU []DisaggregationProtocolStack `json:"du,omitempty"`
	RU []DisaggregationProtocolStack `json:"ru,omitempty"`
}

type NetworkRequirements struct {
	Latency   float32 `json:"latency,omitempty"`
	Bandwidth float32 `json:"bandwidth,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SplitsPlacer{}, &SplitsPlacerList{})
}
