package services

import (
	"context"
	"io"

	"k8s.io/client-go/kubernetes"
)

type Pod struct {
	Name      string
	Namespace string
	Status    PodStatus
	Labels    map[string]string
}

type PodStatus string

const (
	PodRunning   PodStatus = "Running"
	PodPending   PodStatus = "Pending"
	PodFailed    PodStatus = "Failed"
	PodSucceeded PodStatus = "Succeeded"
	PodUnknown   PodStatus = "Unknown"
)

type Context struct {
	Name string
}

type ConfigService interface {
	DiscoverConfig() (string, error)
	LoadContexts(configPath string) ([]Context, string, *kubernetes.Clientset, error)
}

type PodService interface {
	ListPods(ctx context.Context, namespace string) ([]Pod, error)
	SearchPods(ctx context.Context, namespace, query string) ([]Pod, error)
	GetPodLogs(ctx context.Context, podName, namespace string) (io.ReadCloser, error)
}

type Namespace struct {
	Name   string
	Status NamespaceStatus
	Labels map[string]string
}

type NamespaceStatus string

const (
	NamespaceActive      NamespaceStatus = "Active"
	NamespaceTerminating NamespaceStatus = "Terminating"
	NamespaceUnknown     NamespaceStatus = "Unknown"
)

type NamespaceService interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	ListNamespacesDetailed(ctx context.Context) ([]Namespace, error)
}
