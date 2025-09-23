package components

import (
	"context"
	"fmt"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// Mock namespace service for testing namespace list
type namespaceListMockNamespaceService struct {
	namespaces []services.Namespace
	err        error
}

func (m *namespaceListMockNamespaceService) ListNamespaces(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	names := make([]string, len(m.namespaces))
	for i, ns := range m.namespaces {
		names[i] = ns.Name
	}
	return names, nil
}

func (m *namespaceListMockNamespaceService) ListNamespacesDetailed(ctx context.Context) ([]services.Namespace, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.namespaces, nil
}

// Mock pod service for testing namespace list
type namespaceListMockPodService struct{}

func (m *namespaceListMockPodService) ListPods(ctx context.Context, namespace string) ([]services.Pod, error) {
	return []services.Pod{}, nil
}

func (m *namespaceListMockPodService) SearchPods(ctx context.Context, namespace, query string) ([]services.Pod, error) {
	return []services.Pod{}, nil
}

func (m *namespaceListMockPodService) GetPodLogs(ctx context.Context, podName, namespace string) (io.ReadCloser, error) {
	return nil, nil
}

// Mock config service for testing namespace list
type namespaceListMockConfigService struct{}

func (m *namespaceListMockConfigService) DiscoverConfig() (string, error) {
	return "", nil
}

func (m *namespaceListMockConfigService) LoadContexts(configPath string) ([]services.Context, string, *kubernetes.Clientset, error) {
	return []services.Context{{Name: "test-context"}}, "test-context", nil, nil
}

func createTestKubeopticForNamespaceList(namespaces []services.Namespace) *models.Kubeoptic {
	namespaceSvc := &namespaceListMockNamespaceService{namespaces: namespaces}
	podSvc := &namespaceListMockPodService{}
	configSvc := &namespaceListMockConfigService{}

	return models.NewKubeoptic(configSvc, podSvc, namespaceSvc)
}

func TestNamespaceListInitialization(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "kube-system", Status: services.NamespaceActive},
		{Name: "test-ns", Status: services.NamespaceTerminating},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	// Simulate loading namespaces manually since we don't have the full K8s client
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)

	// Test initial state
	if nl == nil {
		t.Fatal("NewNamespaceList returned nil")
	}

	if nl.focused {
		t.Error("Expected namespace list to not be focused initially")
	}

	if nl.kubeoptic != kubeoptic {
		t.Error("Expected kubeoptic reference to be set correctly")
	}
}

func TestNamespaceListLoadNamespaces(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "kube-system", Status: services.NamespaceActive},
		{Name: "test-ns", Status: services.NamespaceTerminating},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)

	// Load namespaces
	cmd := nl.LoadNamespaces()
	// The command might be nil, which is okay

	// Execute the command if it's not nil
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	// Check that items were loaded
	items := nl.list.Items()
	if len(items) != len(namespaces) {
		t.Errorf("Expected %d items, got %d", len(namespaces), len(items))
	}

	// Check that the first item is correct
	if len(items) > 0 {
		item := items[0].(namespaceItem)
		if item.name != "default" {
			t.Errorf("Expected first item name to be 'default', got '%s'", item.name)
		}
		if item.status != "Active" {
			t.Errorf("Expected first item status to be 'Active', got '%s'", item.status)
		}
	}
}

func TestNamespaceListFiltering(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "kube-system", Status: services.NamespaceActive},
		{Name: "kube-node-lease", Status: services.NamespaceActive},
		{Name: "my-app", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.LoadNamespaces()

	// Test that filtering is enabled
	if !nl.list.FilteringEnabled() {
		t.Error("Expected filtering to be enabled")
	}

	// Test search query methods
	nl.SetSearchQuery("kube")
	if nl.GetSearchQuery() != nl.list.FilterValue() {
		t.Error("Search query not set correctly")
	}

	// Test clear search
	nl.ClearSearch()
	if nl.list.FilterValue() != "" {
		t.Error("Filter should be cleared")
	}
}

func TestNamespaceListKeyHandling(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "test-ns", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.Focus() // Set focus to enable key handling
	nl.LoadNamespaces()

	// Test Enter key for selection
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := nl.Update(enterMsg)

	if model == nil {
		t.Error("Update should return a model")
	}

	if cmd == nil {
		t.Error("Enter key should return a command")
	}

	// Test escape key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = nl.Update(escMsg)

	if cmd == nil {
		t.Error("Escape key should return a navigation command")
	}

	// Test refresh key
	refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd = nl.Update(refreshMsg)

	// Refresh might or might not return a command
	_ = cmd
}

func TestNamespaceListFocus(t *testing.T) {
	kubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	nl := NewNamespaceList(kubeoptic)

	// Test initial focus state
	if nl.IsFocused() {
		t.Error("Component should not be focused initially")
	}

	// Test focus
	cmd := nl.Focus()
	if cmd == nil {
		t.Error("Focus should return a command")
	}

	if !nl.IsFocused() {
		t.Error("Component should be focused after Focus() call")
	}

	// Test blur
	cmd = nl.Blur()
	if cmd == nil {
		t.Error("Blur should return a command")
	}

	if nl.IsFocused() {
		t.Error("Component should not be focused after Blur() call")
	}
}

func TestNamespaceListResize(t *testing.T) {
	kubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	nl := NewNamespaceList(kubeoptic)

	// Test resize
	width, height := 100, 50
	nl.SetSize(width, height)

	gotWidth, gotHeight := nl.GetSize()
	if gotWidth != width || gotHeight != height {
		t.Errorf("Expected size %dx%d, got %dx%d", width, height, gotWidth, gotHeight)
	}

	// Test window resize message
	resizeMsg := tea.WindowSizeMsg{Width: 80, Height: 40}
	nl.Update(resizeMsg)

	gotWidth, gotHeight = nl.GetSize()
	if gotWidth != 80 || gotHeight != 40 {
		t.Errorf("Expected size 80x40 after window resize, got %dx%d", gotWidth, gotHeight)
	}
}

func TestNamespaceListStatusMethods(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "test-ns", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.LoadNamespaces()

	// Test status text
	status := nl.GetStatusText()
	if status == "" {
		t.Error("Status text should not be empty")
	}

	// Test status type
	statusType := nl.GetStatusType()
	if statusType != tui.StatusInfo {
		t.Errorf("Expected status type %v, got %v", tui.StatusInfo, statusType)
	}

	// Test with empty namespaces
	emptyKubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	emptyKubeoptic.SetNamespaces([]services.Namespace{})

	emptyNl := NewNamespaceList(emptyKubeoptic)
	emptyNl.LoadNamespaces()

	emptyStatusType := emptyNl.GetStatusType()
	if emptyStatusType != tui.StatusWarning {
		t.Errorf("Expected warning status type for empty list, got %v", emptyStatusType)
	}
}

func TestNamespaceListView(t *testing.T) {
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.SetSize(80, 24)
	nl.LoadNamespaces()

	// Test view rendering
	view := nl.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Test that it shows loading when no size is set
	nlNoSize := NewNamespaceList(kubeoptic)
	viewNoSize := nlNoSize.View()
	if viewNoSize != "Loading namespaces..." {
		t.Errorf("Expected loading message, got: %s", viewNoSize)
	}
}

func TestNamespaceListColorCoding(t *testing.T) {
	tests := []struct {
		name            string
		namespaceStatus services.NamespaceStatus
		expectedStatus  string
	}{
		{"Active namespace", services.NamespaceActive, "Active"},
		{"Terminating namespace", services.NamespaceTerminating, "Terminating"},
		{"Unknown namespace", services.NamespaceUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespaces := []services.Namespace{
				{Name: "test-ns", Status: tt.namespaceStatus},
			}

			kubeoptic := createTestKubeopticForNamespaceList(namespaces)
			kubeoptic.SetNamespaces(namespaces)

			nl := NewNamespaceList(kubeoptic)
			nl.LoadNamespaces()

			items := nl.list.Items()
			if len(items) != 1 {
				t.Fatalf("Expected 1 item, got %d", len(items))
			}

			item := items[0].(namespaceItem)
			if item.status != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, item.status)
			}
		})
	}
}

func TestNamespaceListErrorHandling(t *testing.T) {
	kubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	nl := NewNamespaceList(kubeoptic)

	// Test error state methods - when there are no items, it's not necessarily "loading"
	loading := nl.GetLoadingState()
	_ = loading // loading state depends on implementation

	if err := nl.GetErrorState(); err != nil {
		t.Errorf("Expected no error state initially, got: %v", err)
	}

	// Test error message handling
	errorMsg := tui.ErrorMsg{
		Error:   fmt.Errorf("test error"),
		Context: "testing",
	}

	_, cmd := nl.Update(errorMsg)
	// Should handle error gracefully
	if cmd != nil {
		// Command might be returned for error handling
	}
}

func TestNamespaceListDataProvider(t *testing.T) {
	kubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	nl := NewNamespaceList(kubeoptic)

	// Test refresh data - might or might not return a command
	cmd := nl.RefreshData()
	_ = cmd // Command is optional

	// Test search results
	results := nl.GetSearchResults()
	if results == nil {
		t.Error("GetSearchResults should not return nil")
	}
}

// Benchmark tests for performance
func BenchmarkNamespaceListLoadNamespaces(b *testing.B) {
	// Create a large number of namespaces
	namespaces := make([]services.Namespace, 1000)
	for i := 0; i < 1000; i++ {
		namespaces[i] = services.Namespace{
			Name:   fmt.Sprintf("namespace-%d", i),
			Status: services.NamespaceActive,
		}
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nl.LoadNamespaces()
	}
}

func BenchmarkNamespaceListView(b *testing.B) {
	namespaces := make([]services.Namespace, 100)
	for i := 0; i < 100; i++ {
		namespaces[i] = services.Namespace{
			Name:   fmt.Sprintf("namespace-%d", i),
			Status: services.NamespaceActive,
		}
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.SetSize(80, 24)
	nl.LoadNamespaces()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nl.View()
	}
}
