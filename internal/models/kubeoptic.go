package models

import (
	"context"
	"fmt"
	"io"

	"kubeoptic/internal/services"
)

type ViewType int

const (
	ContextView ViewType = iota
	NamespaceView
	PodView
	LogView
)

type Kubeoptic struct {
	// Services
	configSvc    services.ConfigService
	podSvc       services.PodService
	namespaceSvc services.NamespaceService

	// Navigation state
	focusedView  ViewType
	contexts     []services.Context
	namespaces   []services.Namespace
	pods         []services.Pod
	filteredPods []services.Pod

	// Current selections
	selectedContext   string
	selectedNamespace string
	selectedPod       *services.Pod

	// Search state
	podSearchQuery string
	showingXofY    string

	// Log streaming
	logBuffer   []string
	isFollowing bool
	logStream   io.ReadCloser
}

func NewKubeoptic(configSvc services.ConfigService, podSvc services.PodService, namespaceSvc services.NamespaceService) *Kubeoptic {
	return &Kubeoptic{
		configSvc:         configSvc,
		podSvc:            podSvc,
		namespaceSvc:      namespaceSvc,
		focusedView:       ContextView,
		selectedNamespace: "default",
	}
}

// Navigation methods
func (k *Kubeoptic) SelectContext(contextName string) error {
	for _, ctx := range k.contexts {
		if ctx.Name == contextName {
			k.selectedContext = contextName
			k.focusedView = NamespaceView
			return k.refreshNamespaces()
		}
	}
	return fmt.Errorf("context %s not found", contextName)
}

func (k *Kubeoptic) SelectNamespace(namespace string) error {
	k.selectedNamespace = namespace
	k.focusedView = PodView
	return k.refreshPods()
}

func (k *Kubeoptic) SelectPod(podName string) error {
	for _, pod := range k.filteredPods {
		if pod.Name == podName {
			k.selectedPod = &pod
			k.focusedView = LogView
			return k.startLogStream()
		}
	}
	return fmt.Errorf("pod %s not found", podName)
}

// Search methods
func (k *Kubeoptic) SearchPods(query string) error {
	k.podSearchQuery = query
	ctx := context.Background()

	filteredPods, err := k.podSvc.SearchPods(ctx, k.selectedNamespace, query)
	if err != nil {
		return fmt.Errorf("failed to search pods: %w", err)
	}

	k.filteredPods = filteredPods
	k.updatePodCount()
	return nil
}

func (k *Kubeoptic) ClearSearch() error {
	k.podSearchQuery = ""
	k.filteredPods = k.pods
	k.updatePodCount()
	return nil
}

// Data loading methods
func (k *Kubeoptic) LoadContexts(configPath string) error {
	contexts, currentContext, client, err := k.configSvc.LoadContexts(configPath)
	if err != nil {
		return fmt.Errorf("failed to load contexts: %w", err)
	}

	k.contexts = contexts
	k.selectedContext = currentContext

	// Update services with the new client
	k.podSvc = services.NewPodService(client)
	k.namespaceSvc = services.NewNamespaceService(client)

	return k.refreshNamespaces()
}

func (k *Kubeoptic) refreshNamespaces() error {
	ctx := context.Background()
	namespaces, err := k.namespaceSvc.ListNamespacesDetailed(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh namespaces: %w", err)
	}

	k.namespaces = namespaces
	return nil
}

func (k *Kubeoptic) refreshPods() error {
	ctx := context.Background()
	pods, err := k.podSvc.ListPods(ctx, k.selectedNamespace)
	if err != nil {
		return fmt.Errorf("failed to refresh pods: %w", err)
	}

	k.pods = pods
	k.filteredPods = pods
	k.updatePodCount()
	return nil
}

func (k *Kubeoptic) startLogStream() error {
	if k.selectedPod == nil {
		return fmt.Errorf("no pod selected")
	}

	// Close existing stream
	if k.logStream != nil {
		k.logStream.Close()
	}

	ctx := context.Background()
	stream, err := k.podSvc.GetPodLogs(ctx, k.selectedPod.Name, k.selectedNamespace)
	if err != nil {
		return fmt.Errorf("failed to start log stream: %w", err)
	}

	k.logStream = stream
	k.isFollowing = true
	return nil
}

func (k *Kubeoptic) updatePodCount() {
	if k.podSearchQuery == "" {
		k.showingXofY = fmt.Sprintf("%d pods", len(k.pods))
	} else {
		k.showingXofY = fmt.Sprintf("%d of %d pods", len(k.filteredPods), len(k.pods))
	}
}

// Getters for TUI
func (k *Kubeoptic) GetContexts() []services.Context {
	return k.contexts
}

func (k *Kubeoptic) GetNamespaces() []services.Namespace {
	return k.namespaces
}

func (k *Kubeoptic) GetPods() []services.Pod {
	return k.filteredPods
}

func (k *Kubeoptic) GetSelectedContext() string {
	return k.selectedContext
}

func (k *Kubeoptic) GetSelectedNamespace() string {
	return k.selectedNamespace
}

func (k *Kubeoptic) GetSelectedPod() *services.Pod {
	return k.selectedPod
}

func (k *Kubeoptic) GetFocusedView() ViewType {
	return k.focusedView
}

func (k *Kubeoptic) GetSearchQuery() string {
	return k.podSearchQuery
}

func (k *Kubeoptic) GetPodCount() string {
	return k.showingXofY
}

func (k *Kubeoptic) IsFollowing() bool {
	return k.isFollowing
}

func (k *Kubeoptic) GetLogBuffer() []string {
	return k.logBuffer
}

// SetNamespaces sets the namespaces list directly (used for testing)
func (k *Kubeoptic) SetNamespaces(namespaces []services.Namespace) {
	k.namespaces = namespaces
}
