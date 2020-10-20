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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	topologyKey = "topology"

	logNodeKey  = "node"
	logSplitKey = "split"
)

// SplitsPlacerReconciler reconciles a SplitsPlacer object
type SplitsPlacerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=oai.unisinos,resources=splitsplacers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oai.unisinos,resources=splitsplacers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get

func (r *SplitsPlacerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("splitsplacer", req.NamespacedName)

	splitsPlacer := &oaiv1beta1.SplitsPlacer{}
	if err := r.Get(ctx, req.NamespacedName, splitsPlacer); err != nil {
		log.Error(err, "unable to fetch SplitsPlacer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.syncTopology(splitsPlacer, log); err != nil {
		return ctrl.Result{}, fmt.Errorf("error syncing topology: %s", err)
	}

	if err := r.syncSplits(splitsPlacer, log); err != nil {
		return ctrl.Result{}, fmt.Errorf("error syncing splits: %w", err)
	}

	return ctrl.Result{Requeue: true, RequeueAfter: resyncPeriod}, nil
}

func (r *SplitsPlacerReconciler) syncTopology(splitsPlacer *oaiv1beta1.SplitsPlacer, log logr.Logger) error {
	if splitsPlacer.Spec.TopologyConfig == "" {
		log.Info("topology not provided, skipping sync")
		return nil
	}

	topology := &oaiv1beta1.Topology{}
	topologyKey := r.getObjectKey(splitsPlacer.Spec.TopologyConfig, splitsPlacer.Namespace)
	if err := r.readTopology(topologyKey, topology); err != nil {
		return fmt.Errorf("error reading topology: %w", err)
	}

	if errors := r.validateTopologyNodes(topology, splitsPlacer.Namespace, log); errors != nil {
		for _, err := range errors {
			r.Recorder.Event(splitsPlacer, EventErrorType, "InvalidTopologyNode", err.Error())
		}
		return fmt.Errorf("error validating topology nodes")
	}

	return nil
}

func (r *SplitsPlacerReconciler) validateTopologyNodes(topology *oaiv1beta1.Topology, namespace string,
	log logr.Logger) []error {

	var errorPool []error
	for _, node := range topology.Nodes {
		k8sNode := &v1.Node{}
		nodeKey := r.getObjectKey(node.Name, namespace)
		if exists, err := GetNode(r.Client, nodeKey, k8sNode); err != nil {
			log.Error(err, "error getting node", logNodeKey, nodeKey.Name)
			errorPool = append(errorPool, fmt.Errorf("error getting node '%s': %w", nodeKey.Name, err))
		} else if !exists {
			errorPool = append(errorPool, fmt.Errorf("node '%s' described in topology not found", nodeKey.Name))
		}
	}

	if len(errorPool) > 0 {
		return errorPool
	}

	return nil
}

func (r *SplitsPlacerReconciler) syncSplits(splitsPlacer *oaiv1beta1.SplitsPlacer, log logr.Logger) error {
	for _, ru := range splitsPlacer.Spec.RUs {
		// Check if split exists
		splitKey := r.getObjectKey(ru.SplitName, splitsPlacer.Namespace)

		split := &oaiv1beta1.Split{}
		exists, err := GetSplit(r.Client, splitKey, split)
		if err != nil {
			return fmt.Errorf("error checking if split exists: %w", err)
		}

		if !exists {
			log.Info("Creating split...", logSplitKey, ru.SplitName)
			err := r.Create(context.Background(), r.getSplitTemplate(ru, splitsPlacer.Namespace,
				splitsPlacer.Spec.CoreIP))
			if err != nil {
				return fmt.Errorf("error creating split %s: %w", ru.SplitName, err)
			}
		}

		log.Info("Split already exists, skipping creation...", logSplitKey, ru.SplitName)
	}

	return nil
}

func (r *SplitsPlacerReconciler) getSplitTemplate(ru oaiv1beta1.RUPosition, namespace string,
	coreIP string) *oaiv1beta1.Split {
	return &oaiv1beta1.Split{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ru.SplitName,
			Namespace: namespace,
		},
		Spec: oaiv1beta1.SplitSpec{
			CoreIP: coreIP,
			RUNode: ru.Node,
		},
	}
}

func (r *SplitsPlacerReconciler) readTopology(objectKey types.NamespacedName,
	topology *oaiv1beta1.Topology) error {
	cm := &v1.ConfigMap{}
	if exists, err := GetConfigMap(r.Client, objectKey, cm); err != nil {
		return fmt.Errorf("error getting topology '%s' config map: %w", objectKey.String(), err)
	} else if !exists {
		return fmt.Errorf("topology config map '%s' does not exists", objectKey.String())
	}

	topologyData, exists := cm.BinaryData[topologyKey]
	if !exists {
		return fmt.Errorf("invalid topology config map. Key '%s' does not exist", topologyKey)
	}

	if err := json.Unmarshal(topologyData, topology); err != nil {
		return fmt.Errorf("error unmarshaling topology: %w", err)
	}

	return nil
}

func (r *SplitsPlacerReconciler) getObjectKey(name string,
	namespace string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func (r *SplitsPlacerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oaiv1beta1.SplitsPlacer{}).
		Complete(r)
}
