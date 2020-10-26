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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	"github.com/juliorenner/oai-k8s/operator/controllers/algorithm"
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
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
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

func (r *SplitsPlacerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("splitsplacer", req.NamespacedName)

	splitsPlacer := &oaiv1beta1.SplitsPlacer{}
	if err := r.Get(ctx, req.NamespacedName, splitsPlacer); err != nil {
		log.Error(err, "unable to fetch SplitsPlacer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if splitsPlacer.Status.State == oaiv1beta1.PlacerStateFinished && !splitsPlacer.Spec.Retrigger {
		log.Info("Skipping reconcile. Status Finished and retrigger not enabled.", logSplitKey, splitsPlacer.Name)
		return ctrl.Result{}, nil
	}

	if err := r.syncTopology(splitsPlacer, log); err != nil {
		if err := r.updateStatus(splitsPlacer, oaiv1beta1.PlacerStateError); err != nil {
			log.Error(err, "error updating splits placer status after error syncing topology")
		}
		return ctrl.Result{}, fmt.Errorf("error syncing topology: %s", err)
	}

	if err := r.syncSplits(splitsPlacer, log); err != nil {
		if err := r.updateStatus(splitsPlacer, oaiv1beta1.PlacerStateError); err != nil {
			log.Error(err, "error updating splits placer status after error syncing splits")
		}
		return ctrl.Result{}, fmt.Errorf("error syncing splits: %w", err)
	}

	r.Recorder.Event(splitsPlacer, v1.EventTypeNormal, "Sync", "Synced successfully")

	if err := r.Update(context.Background(), splitsPlacer); err != nil {
		log.Error(err, "error updating splits placer spec")
		return ctrl.Result{}, fmt.Errorf("error updating splits placer: %w", err)
	}

	if err := r.updateStatus(splitsPlacer, oaiv1beta1.PlacerStateFinished); err != nil {
		log.Error(err, "error updating splits placer status")
	}

	return ctrl.Result{}, nil
}

func (r *SplitsPlacerReconciler) updateStatus(splitsPlacer *oaiv1beta1.SplitsPlacer, desiredState oaiv1beta1.SplitsPlacerState) error {
	if splitsPlacer.Status.State != desiredState {
		splitsPlacer.Status.State = desiredState
		if err := r.Status().Update(context.Background(), splitsPlacer); err != nil {
			return fmt.Errorf("error updating splitsplacer status: %w", err)
		}
	}

	return nil
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

	if topologyErrors := r.validateTopologyNodes(topology, log); topologyErrors != nil {
		for _, err := range topologyErrors {
			r.Recorder.Event(splitsPlacer, v1.EventTypeWarning, "InvalidTopologyNode", err.Error())
		}
		return errors.New("error validating topology nodes")
	}

	disaggregation := map[string]*oaiv1beta1.Disaggregation{}
	if err := r.readDisaggregationsMetadata(disaggregation); err != nil {
		return fmt.Errorf("error reading disaggregation metadata: %w", err)
	}

	if err := r.place(splitsPlacer, topology, disaggregation, log); err != nil {
		return fmt.Errorf("error placing service functions: %w", err)
	}

	return nil
}

func (r *SplitsPlacerReconciler) validateTopologyNodes(topology *oaiv1beta1.Topology, log logr.Logger) []error {
	var errorPool []error

	nodeList := &v1.NodeList{}
	if err := utils.ListNodes(r.Client, nodeList); err != nil {
		return append(errorPool, err)
	}

	k8sNodeMap := utils.NodeListToMap(nodeList)
	for nodeName := range topology.Nodes {
		if _, exists := k8sNodeMap[nodeName]; !exists {
			errorPool = append(errorPool, fmt.Errorf("node '%s' does not exist", nodeName))
			continue
		} else if len(errorPool) > 0 {
			// if there are errors skip resources assignment and finish the validation
			continue
		}
	}

	if len(errorPool) > 0 {
		return errorPool
	}

	log.Info("topology successfully validated")
	return nil
}

func (r *SplitsPlacerReconciler) syncSplits(splitsPlacer *oaiv1beta1.SplitsPlacer, log logr.Logger) error {
	for _, ru := range splitsPlacer.Spec.RUs {
		// Check if split exists
		splitKey := r.getObjectKey(ru.SplitName, splitsPlacer.Namespace)

		split := &oaiv1beta1.Split{}
		exists, err := utils.GetSplit(r.Client, splitKey, split)
		if err != nil {
			return fmt.Errorf("error checking if split exists: %w", err)
		}

		if !exists {
			log.Info("Creating split...", logSplitKey, ru.SplitName)
			split = r.getSplitTemplate(ru, splitsPlacer.Namespace, splitsPlacer.Spec.CoreIP)
			if err := ctrl.SetControllerReference(splitsPlacer, split, r.Scheme); err != nil {
				return fmt.Errorf("error setting split owner reference: %w", err)
			}
			err := r.Create(context.Background(), split)
			if err != nil {
				return fmt.Errorf("error creating split %s: %w", ru.SplitName, err)
			}
		}

		log.Info("Split already exists, skipping creation...", logSplitKey, ru.SplitName)
	}

	return nil
}

func (r *SplitsPlacerReconciler) getSplitTemplate(ru *oaiv1beta1.RUPosition, namespace string,
	coreIP string) *oaiv1beta1.Split {
	return &oaiv1beta1.Split{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ru.SplitName,
			Namespace: namespace,
		},
		Spec: oaiv1beta1.SplitSpec{
			CoreIP: coreIP,
			RUNode: ru.RUNode,
			CUNode: ru.CUNode,
			DUNode: ru.DUNode,
		},
	}
}

func (r *SplitsPlacerReconciler) readDisaggregationsMetadata(disaggregation map[string]*oaiv1beta1.
	Disaggregation) error {
	cmObjectKey := types.NamespacedName{
		Namespace: operatorNamespace,
		Name:      DisaggregationConfigMapName,
	}

	cm := &v1.ConfigMap{}
	if exists, err := utils.GetConfigMap(r.Client, cmObjectKey, cm); err != nil {
		return fmt.Errorf("error getting disaggregation config map: %w", err)
	} else if !exists {
		return fmt.Errorf("disaggregation config map '%s' not found in namespace '%s'", cmObjectKey.Name,
			cmObjectKey.Namespace)
	}

	disaggregationInt := make(map[string]interface{})
	disaggregationData := []byte(cm.Data[DisaggregationKey])
	if err := json.Unmarshal(disaggregationData, &disaggregationInt); err != nil {
		return fmt.Errorf("error unmarshaling disaggregation config map data: %w. Data: %s", err, cm.Data[DisaggregationKey])
	}

	for k, v := range disaggregationInt {
		d, ok := v.(*oaiv1beta1.Disaggregation)
		if !ok {
			return fmt.Errorf("error casting dissagregation interface to struct object. Key: %s", k)
		}
		disaggregation[k] = d
	}

	return nil
}

func (r *SplitsPlacerReconciler) readTopology(objectKey types.NamespacedName, topology *oaiv1beta1.Topology) error {
	cm := &v1.ConfigMap{}
	if exists, err := utils.GetConfigMap(r.Client, objectKey, cm); err != nil {
		return fmt.Errorf("error getting topology '%s' config map: %w", objectKey.String(), err)
	} else if !exists {
		return fmt.Errorf("topology config map '%s' does not exists", objectKey.String())
	}

	topologyData, exists := cm.Data[topologyKey]
	if !exists {
		return fmt.Errorf("invalid topology config map. Key '%s' does not exist", topologyKey)
	}

	if err := json.Unmarshal([]byte(topologyData), topology); err != nil {
		return fmt.Errorf("error unmarshaling topology: %w", err)
	}

	return nil
}

func (r *SplitsPlacerReconciler) getObjectKey(name string, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func (r *SplitsPlacerReconciler) place(splitsPlacer *oaiv1beta1.SplitsPlacer, topology *oaiv1beta1.Topology,
	disaggregations map[string]*oaiv1beta1.Disaggregation, log logr.Logger) error {

	nodeList := &v1.NodeList{}
	if err := utils.ListNodes(r.Client, nodeList); err != nil {
		return fmt.Errorf("error listing K8S nodes: %w", err)
	}

	requestedResources := &utils.RequestedResources{
		Memory: *utils.NewMemoryQuantity(SplitMemoryRequestValue),
		CPU:    *utils.NewCPUQuantity(SplitCPURequestValue),
	}

	topologyGraph := algorithm.NewPlacementBFS(topology, disaggregations, nodeList, requestedResources, log)

	if success, err := topologyGraph.Place(splitsPlacer.Spec.RUs); err != nil {
		return err
	} else if !success {
		return errors.New("not possible to allocate all RUs")
	}

	return nil
}

func (r *SplitsPlacerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oaiv1beta1.SplitsPlacer{}).
		Complete(r)
}
