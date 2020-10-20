package controllers

import (
	"context"
	"fmt"

	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSplit(k8sClient client.Client, objectKey types.NamespacedName,
	split *oaiv1beta1.Split) (bool, error) {
	if err := k8sClient.Get(context.Background(), objectKey, split); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("error getting split: %w", err)
	}

	return true, nil
}

// TODO: Use Informer/Cache
func GetDeployment(k8sClient client.Client, objectKey types.NamespacedName,
	deployment *appsv1.Deployment) (bool, error) {
	err := k8sClient.Get(context.Background(), objectKey, deployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("error getting deployment %s: %w", objectKey.String(), err)
	}

	return true, nil
}

// TODO: Use Informer/Cache
func GetConfigMap(k8sClient client.Client, objectKey types.NamespacedName, cm *v1.ConfigMap) (bool, error) {
	err := k8sClient.Get(context.Background(), objectKey, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("error getting config map %s: %w", objectKey.String(), err)
	}

	return true, nil
}

// TODO: Use Informer/Cache
func ListNodes(k8sClient client.Client, nodeList *v1.NodeList) error {
	err := k8sClient.List(context.Background(), nodeList)
	if err != nil {
		return fmt.Errorf("error listing nodes: %w", err)
	}

	return nil
}
