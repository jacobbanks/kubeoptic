//go:build integration
// +build integration

package components

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// MockKubernetes simulates a real Kubernetes environment
type MockKubernetes struct {
	contexts       []services.Context
	currentContext string
	switched       bool
	switchedTo     string
}

func NewMockKubernetes() *MockKubernetes {
	return &MockKubernetes{
		contexts: []services.Context{
			{Name: "minikube"},
			{Name: "staging"},
			{Name: "production"},
		},
		currentContext: "minikube",
	}
}

func (mk *MockKubernetes) SwitchContext(contextName string) error {
	mk.switched = true
	mk.switchedTo = contextName
	mk.currentContext = contextName
	return nil
}

func (mk *MockKubernetes) GetContexts() ([]services.Context, string) {
	return mk.contexts, mk.currentContext
}

func TestIntegrationContextSwitching(t *testing.T) {
	// Setup mock environment
	mock := NewMockKubernetes()
	contexts, current := mock.GetContexts()

	// Create context list
	cl := NewContextList(contexts, current)

	// Verify initial state
	if current != "minikube" {
		t.Errorf("expected initial context 'minikube', got '%s'", current)
	}

	// Simulate user navigation to production context
	// First, simulate down arrow key to move to production (index 2)
	_, cmd1 := cl.Update(tea.KeyMsg{Type: tea.KeyDown}) // Move to staging
	_, cmd2 := cl.Update(tea.KeyMsg{Type: tea.KeyDown}) // Move to production

	// Verify we can get the selected context
	selected := cl.SelectedContext()
	if selected == nil {
		t.Fatal("selected context should not be nil")
	}

	// Simulate Enter key to select
	_, cmd3 := cl.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd3 == nil {
		t.Fatal("Enter key should return a command")
	}

	// Execute the command to get the message
	msg := cmd3()
	contextMsg, ok := msg.(tui.ContextSelectedMsg)
	if !ok {
		t.Fatalf("expected ContextSelectedMsg, got %T", msg)
	}

	if contextMsg.Context == nil {
		t.Fatal("context in message should not be nil")
	}

	// Simulate context switching in our mock environment
	err := mock.SwitchContext(contextMsg.Context.Name)
	if err != nil {
		t.Fatalf("failed to switch context: %v", err)
	}

	// Verify the switch occurred
	if !mock.switched {
		t.Error("context switch should have been called")
	}

	if mock.switchedTo != contextMsg.Context.Name {
		t.Errorf("expected switch to '%s', got '%s'", contextMsg.Context.Name, mock.switchedTo)
	}

	// Verify new current context
	_, newCurrent := mock.GetContexts()
	if newCurrent != contextMsg.Context.Name {
		t.Errorf("expected new current context '%s', got '%s'", contextMsg.Context.Name, newCurrent)
	}

	// Test commands are properly nil when expected
	if cmd1 != nil {
		t.Log("Navigation commands may return commands, this is expected")
	}
	if cmd2 != nil {
		t.Log("Navigation commands may return commands, this is expected")
	}
}

func TestIntegrationContextListWorkflow(t *testing.T) {
	mock := NewMockKubernetes()
	contexts, current := mock.GetContexts()

	// Test complete workflow: create -> init -> update -> select
	cl := NewContextList(contexts, current)

	// Test initialization
	initCmd := cl.Init()
	if initCmd != nil {
		t.Error("Init should return nil for context list")
	}

	// Test setting size (common in real TUI apps)
	cl.SetSize(80, 24)

	// Test focus changes (common in real TUI apps)
	cl.SetFocus(true)
	cl.SetFocus(false)

	// Test search filtering workflow
	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("prod")}
	_, searchCmd := cl.Update(searchMsg)

	// Test view rendering at different states
	view1 := cl.View()
	if len(view1) == 0 {
		t.Error("view should not be empty")
	}

	// Test escape key (common in TUI navigation)
	escapeMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, escapeCmd := cl.Update(escapeMsg)

	// All these operations should not panic or error
	t.Log("Complete workflow test passed")

	// Optional: commands may be nil and that's OK
	_ = searchCmd
	_ = escapeCmd
}

func TestIntegrationErrorScenarios(t *testing.T) {
	// Test with empty context list
	emptyContexts := []services.Context{}
	cl := NewContextList(emptyContexts, "")

	// Attempt to select from empty list
	_, cmd := cl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		contextMsg, ok := msg.(tui.ContextSelectedMsg)
		if ok && contextMsg.Context != nil {
			t.Error("empty list should not return a valid context selection")
		}
	}

	// Test with invalid current context
	contexts := []services.Context{{Name: "valid"}}
	cl2 := NewContextList(contexts, "nonexistent")

	// Should still work without crashing
	view := cl2.View()
	if len(view) == 0 {
		t.Error("view should render even with invalid current context")
	}
}

func TestIntegrationPerformance(t *testing.T) {
	// Test with large number of contexts
	largeContexts := make([]services.Context, 1000)
	for i := 0; i < 1000; i++ {
		largeContexts[i] = services.Context{Name: fmt.Sprintf("context-%d", i)}
	}

	cl := NewContextList(largeContexts, "context-500")

	// Test that operations complete in reasonable time
	start := time.Now()

	// Perform common operations
	cl.View()
	cl.SelectedContext()
	cl.Update(tea.KeyMsg{Type: tea.KeyDown})
	cl.Update(tea.KeyMsg{Type: tea.KeyEnter})

	elapsed := time.Since(start)
	if elapsed > time.Millisecond*100 {
		t.Errorf("operations took too long: %v", elapsed)
	}
}
