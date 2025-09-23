//go:build integration
// +build integration

package components

import (
	"context"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// Integration tests for namespace list component
// These tests simulate real user workflows and interactions

func TestNamespaceListIntegrationWorkflow(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test data that simulates a real cluster
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "kube-system", Status: services.NamespaceActive},
		{Name: "kube-public", Status: services.NamespaceActive},
		{Name: "my-app-prod", Status: services.NamespaceActive},
		{Name: "my-app-staging", Status: services.NamespaceActive},
		{Name: "cleanup-ns", Status: services.NamespaceTerminating},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)

	// Test complete workflow: Initialize -> Load -> Focus -> Navigate -> Select

	// 1. Initialize component
	initCmd := nl.Init()
	if initCmd != nil {
		msg := initCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	// 2. Set size (simulates real TUI environment)
	nl.SetSize(80, 24)
	resizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	nl.Update(resizeMsg)

	// 3. Focus the component
	focusCmd := nl.Focus()
	if focusCmd != nil {
		msg := focusCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	// Verify component is ready
	if !nl.IsFocused() {
		t.Error("Component should be focused")
	}

	items := nl.list.Items()
	if len(items) != len(namespaces) {
		t.Errorf("Expected %d namespaces loaded, got %d", len(namespaces), len(items))
	}

	// 4. Test navigation (down arrow key)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	nl.Update(downMsg)

	// 5. Test filtering workflow
	filterMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	nl.Update(filterMsg)

	// Simulate typing a filter
	searchQuery := "kube"
	for _, char := range searchQuery {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		nl.Update(charMsg)
	}

	// 6. Clear filter
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	nl.Update(escMsg)

	// 7. Test namespace selection
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := nl.Update(enterMsg)

	if model == nil {
		t.Error("Update should return the model")
	}

	if cmd == nil {
		t.Error("Namespace selection should trigger an action")
	}

	// Execute the selection command
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			// This would typically trigger a namespace change in the kubeoptic model
			nl.Update(msg)
		}
	}

	// 8. Test refresh functionality
	refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	nl.Update(refreshMsg)

	// 9. Test view rendering under different states
	view := nl.View()
	if view == "" {
		t.Error("View should render content")
	}

	if len(view) < 50 { // Expect substantial content
		t.Errorf("View seems too short, might be missing content: %s", view)
	}
}

func TestNamespaceListIntegrationErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test error handling in realistic scenarios
	kubeoptic := createTestKubeopticForNamespaceList([]services.Namespace{})
	nl := NewNamespaceList(kubeoptic)

	// 1. Test with empty namespace list (cluster with no accessible namespaces)
	nl.SetSize(80, 24)
	nl.Focus()

	view := nl.View()
	if view == "" {
		t.Error("Should render something even with empty namespaces")
	}

	// 2. Test error message handling
	errorMsg := tui.ErrorMsg{
		Error:   context.DeadlineExceeded,
		Context: "fetching namespaces",
	}

	nl.Update(errorMsg)

	// Component should handle errors gracefully and continue to be usable
	if !nl.IsFocused() {
		t.Error("Component should remain focused after error")
	}

	// 3. Test namespace loading after error
	// Simulate recovery by adding namespaces
	namespaces := []services.Namespace{
		{Name: "recovered-ns", Status: services.NamespaceActive},
	}

	kubeoptic.SetNamespaces(namespaces)
	loadCmd := nl.LoadNamespaces()
	if loadCmd != nil {
		msg := loadCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	items := nl.list.Items()
	if len(items) != 1 {
		t.Errorf("Expected 1 namespace after recovery, got %d", len(items))
	}
}

func TestNamespaceListIntegrationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with a large number of namespaces (simulates large cluster)
	namespaces := make([]services.Namespace, 500)
	for i := 0; i < 500; i++ {
		status := services.NamespaceActive
		if i%50 == 0 { // Some terminating namespaces
			status = services.NamespaceTerminating
		}

		namespaces[i] = services.Namespace{
			Name:   generateNamespaceName(i),
			Status: status,
		}
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.SetSize(80, 24)

	// Measure performance of loading large number of namespaces
	start := time.Now()

	loadCmd := nl.LoadNamespaces()
	if loadCmd != nil {
		msg := loadCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	loadTime := time.Since(start)

	// Should load quickly even with many namespaces
	if loadTime > time.Millisecond*100 {
		t.Errorf("Loading %d namespaces took too long: %v", len(namespaces), loadTime)
	}

	// Test rendering performance
	start = time.Now()
	view := nl.View()
	renderTime := time.Since(start)

	if renderTime > time.Millisecond*50 {
		t.Errorf("Rendering view took too long: %v", renderTime)
	}

	if view == "" {
		t.Error("Should render content even with many namespaces")
	}

	// Test filtering performance with large dataset
	nl.Focus()
	start = time.Now()

	// Simulate typing a filter
	filterMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	nl.Update(filterMsg)

	searchQuery := "namespace-1"
	for _, char := range searchQuery {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		nl.Update(charMsg)
	}

	filterTime := time.Since(start)

	if filterTime > time.Millisecond*50 {
		t.Errorf("Filtering took too long: %v", filterTime)
	}
}

func TestNamespaceListIntegrationStateTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test state transitions during namespace lifecycle
	namespaces := []services.Namespace{
		{Name: "test-ns", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	nl := NewNamespaceList(kubeoptic)
	nl.SetSize(80, 24)
	nl.Focus()

	// Initial state - namespace is active
	loadCmd := nl.LoadNamespaces()
	if loadCmd != nil {
		msg := loadCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	items := nl.list.Items()
	if len(items) != 1 {
		t.Fatalf("Expected 1 namespace, got %d", len(items))
	}

	item := items[0].(namespaceItem)
	if item.status != "Active" {
		t.Errorf("Expected Active status, got %s", item.status)
	}

	// Simulate namespace entering terminating state
	namespaces[0].Status = services.NamespaceTerminating
	kubeoptic.SetNamespaces(namespaces)

	// Refresh to pick up status change
	refreshCmd := nl.RefreshData()
	if refreshCmd != nil {
		msg := refreshCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	items = nl.list.Items()
	if len(items) != 1 {
		t.Fatalf("Expected 1 namespace after refresh, got %d", len(items))
	}

	item = items[0].(namespaceItem)
	if item.status != "Terminating" {
		t.Errorf("Expected Terminating status after refresh, got %s", item.status)
	}

	// Simulate namespace being deleted (removed from list)
	kubeoptic.SetNamespaces([]services.Namespace{})

	refreshCmd = nl.RefreshData()
	if refreshCmd != nil {
		msg := refreshCmd()
		if msg != nil {
			nl.Update(msg)
		}
	}

	items = nl.list.Items()
	if len(items) != 0 {
		t.Errorf("Expected 0 namespaces after deletion, got %d", len(items))
	}
}

// Helper function to generate namespace names
func generateNamespaceName(i int) string {
	prefixes := []string{"app", "service", "data", "ml", "web", "api", "cache", "queue"}
	environments := []string{"dev", "staging", "prod", "test"}

	prefix := prefixes[i%len(prefixes)]
	env := environments[(i/len(prefixes))%len(environments)]

	return fmt.Sprintf("%s-%s-%d", prefix, env, i)
}

// Integration test for full kubeoptic integration
func TestNamespaceListKubeopticIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test integration with kubeoptic model state changes
	namespaces := []services.Namespace{
		{Name: "default", Status: services.NamespaceActive},
		{Name: "production", Status: services.NamespaceActive},
	}

	kubeoptic := createTestKubeopticForNamespaceList(namespaces)
	kubeoptic.SetNamespaces(namespaces)

	// Set initial context
	originalContext := "test-context"

	nl := NewNamespaceList(kubeoptic)
	nl.SetSize(80, 24)
	nl.Focus()

	// Load initial namespaces
	nl.LoadNamespaces()

	// Verify context information is displayed
	view := nl.View()
	if !contains(view, originalContext) {
		t.Errorf("View should contain context information: %s", originalContext)
	}

	// Test namespace selection affecting kubeoptic state
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := nl.Update(enterMsg)

	if cmd != nil {
		msg := cmd()
		if msg != nil {
			// This should trigger a namespace selection in kubeoptic
			switch msg := msg.(type) {
			case tui.NamespacesLoadedMsg:
				// Expected message type for namespace operations
				if msg.Error != nil {
					t.Errorf("Namespace selection should not produce error: %v", msg.Error)
				}
			}
		}
	}

	// Test that namespace list responds to external kubeoptic state changes
	newNamespaces := []services.Namespace{
		{Name: "new-namespace", Status: services.NamespaceActive},
	}

	kubeoptic.SetNamespaces(newNamespaces)

	// Send a message to simulate external refresh
	refreshMsg := tui.NamespacesLoadedMsg{
		Namespaces: []string{"new-namespace"},
		Error:      nil,
	}

	nl.Update(refreshMsg)

	// Verify the list was updated
	items := nl.list.Items()
	if len(items) != 1 {
		t.Errorf("Expected 1 namespace after external update, got %d", len(items))
	}

	if len(items) > 0 {
		item := items[0].(namespaceItem)
		if item.name != "new-namespace" {
			t.Errorf("Expected 'new-namespace', got %s", item.name)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
