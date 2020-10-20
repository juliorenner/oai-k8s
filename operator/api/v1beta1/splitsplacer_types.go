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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SplitsPlacerSpec defines the desired state of SplitsPlacer
type SplitsPlacerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// RUs
	// +kubebuilder:validation:Required
	RUs []RUPosition `json:"rus,omitempty"`
	// CoreIP to where the splits created will point to.
	// +kubebuilder:validation:Required
	CoreIP string `json:"coreIP,omitempty"`
	// Topology refers to the config map name where the topology is described
	TopologyConfig string `json:"topologyConfig,omitempty"`
}

// RUPosition defines the position and the name of the RU from one service chain. Based on this definition a Split
// will be created.
type RUPosition struct {
	SplitName string `json:"splitName,omitempty"`
	Node      string `json:"node,omitempty"`
}

// SplitsPlacerStatus defines the observed state of SplitsPlacer
type SplitsPlacerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
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
	Nodes []Node `json:"nodes,omitempty"`
	Links []Link `json:"links,omitempty"`
}

type Node struct {
	Name       string   `json:"name,omitempty"`
	Interfaces []string `json:"interfaces,omitempty"`
	Core       bool     `json:"core,omitempty"`
}

type Link struct {
	Name         string     `json:"name,omitempty"`
	LinkCapacity int        `json:"linkCapacity,omitempty"`
	LinkDelay    float32    `json:"linkDelay,omitempty"`
	Source       Connection `json:"source,omitempty"`
	Destination  Connection `json:"destination,omitempty"`
}

type Connection struct {
	Node      string `json:"node,omitempty"`
	Interface string `json:"interface,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SplitsPlacer{}, &SplitsPlacerList{})
}
