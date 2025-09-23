package services

import (
	"context"
	"fmt"

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