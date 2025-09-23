package services

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceServiceImpl struct {
	client kubernetes.Interface
}

func NewNamespaceService(client kubernetes.Interface) NamespaceService {
	return &NamespaceServiceImpl{
		client: client,
	}
}

func (n *NamespaceServiceImpl) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaceList, err := n.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

func (n *NamespaceServiceImpl) ListNamespacesDetailed(ctx context.Context) ([]Namespace, error) {
	namespaceList, err := n.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]Namespace, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		status := convertNamespaceStatus(ns.Status.Phase)
		namespaces = append(namespaces, Namespace{
			Name:   ns.Name,
			Status: status,
			Labels: ns.Labels,
		})
	}

	return namespaces, nil
}

func convertNamespaceStatus(phase corev1.NamespacePhase) NamespaceStatus {
	switch phase {
	case corev1.NamespaceActive:
		return NamespaceActive
	case corev1.NamespaceTerminating:
		return NamespaceTerminating
	default:
		return NamespaceUnknown
	}
}
