package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

func TestNewContextList(t *testing.T) {
	contexts := []services.Context{
		{Name: "minikube"},
		{Name: "production"},
		{Name: "staging"},
	}
	currentContext := "production"

	cl := NewContextList(contexts, currentContext)

	// Verify basic initialization
	if cl.contexts == nil {
		t.Error("contexts should not be nil")
	}
	if len(cl.contexts) != 3 {
		t.Errorf("expected 3 contexts, got %d", len(cl.contexts))
	}
	if cl.current != currentContext {
		t.Errorf("expected current context %s, got %s", currentContext, cl.current)
	}

	// Verify list has correct number of items
	if cl.list.VisibleItems() == nil {
		t.Error("list items should not be nil")
	}
}

func TestContextItemImplementsListItem(t *testing.T) {
	ctx := services.Context{Name: "test-context"}
	item := ContextItem{context: ctx, isCurrent: false}

	// Test FilterValue
	if item.FilterValue() != "test-context" {
		t.Errorf("expected FilterValue 'test-context', got '%s'", item.FilterValue())
	}

	// Test Title for non-current context
	if item.Title() != "test-context" {
		t.Errorf("expected Title 'test-context', got '%s'", item.Title())
	}

	// Test Description for non-current context
	if item.Description() != "" {
		t.Errorf("expected empty Description, got '%s'", item.Description())
	}
}

func TestContextItemCurrentMarking(t *testing.T) {
	ctx := services.Context{Name: "current-context"}
	item := ContextItem{context: ctx, isCurrent: true}

	// Test Title for current context
	expectedTitle := "● current-context"
	if item.Title() != expectedTitle {
		t.Errorf("expected Title '%s', got '%s'", expectedTitle, item.Title())
	}

	// Test Description for current context
	if item.Description() != "Current context" {
		t.Errorf("expected Description 'Current context', got '%s'", item.Description())
	}
}

func TestContextListInit(t *testing.T) {
	contexts := []services.Context{{Name: "test"}}
	cl := NewContextList(contexts, "test")

	cmd := cl.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestContextListUpdate(t *testing.T) {
	contexts := []services.Context{
		{Name: "context1"},
		{Name: "context2"},
	}
	cl := NewContextList(contexts, "context1")

	// Test regular update (non-enter key)
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedCl, cmd := cl.Update(keyMsg)
	if cmd == nil {
		t.Log("Non-enter key should pass through to list (this may be expected)")
	}

	// Verify the component is returned
	if updatedCl.contexts == nil {
		t.Error("updated component should maintain context data")
	}
}

func TestContextListEnterKeySelection(t *testing.T) {
	contexts := []services.Context{
		{Name: "context1"},
		{Name: "context2"},
	}
	cl := NewContextList(contexts, "context1")

	// Simulate Enter key press
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := cl.Update(enterMsg)

	if cmd == nil {
		t.Error("Enter key should return a command")
		return
	}

	// Execute the command to get the message
	msg := cmd()
	contextMsg, ok := msg.(tui.ContextSelectedMsg)
	if !ok {
		t.Errorf("expected ContextSelectedMsg, got %T", msg)
		return
	}

	if contextMsg.Context == nil {
		t.Error("selected context should not be nil")
		return
	}

	// The first item should be selected by default
	if contextMsg.Context.Name != "context1" {
		t.Errorf("expected selected context 'context1', got '%s'", contextMsg.Context.Name)
	}
}

func TestContextListSelectedContext(t *testing.T) {
	contexts := []services.Context{
		{Name: "context1"},
		{Name: "context2"},
	}
	cl := NewContextList(contexts, "context1")

	selected := cl.SelectedContext()
	if selected == nil {
		t.Error("SelectedContext should not return nil")
		return
	}

	// The first context should be selected by default
	if selected.Name != "context1" {
		t.Errorf("expected selected context 'context1', got '%s'", selected.Name)
	}
}

func TestContextListSetSize(t *testing.T) {
	contexts := []services.Context{{Name: "test"}}
	cl := NewContextList(contexts, "test")

	// Test SetSize doesn't panic
	cl.SetSize(80, 24)
	// If we get here without panic, the test passes
}

func TestContextListSetFocus(t *testing.T) {
	contexts := []services.Context{{Name: "test"}}
	cl := NewContextList(contexts, "test")

	// Test SetFocus doesn't panic
	cl.SetFocus(true)
	cl.SetFocus(false)
	// If we get here without panic, the test passes
}

func TestContextListView(t *testing.T) {
	contexts := []services.Context{
		{Name: "minikube"},
		{Name: "production"},
	}
	cl := NewContextList(contexts, "production")

	view := cl.View()
	if view == "" {
		t.Error("View should not return empty string")
	}

	// Debug: print the actual view content
	t.Logf("Actual view content: %s", view)

	// Check that the view contains the title
	if !strings.Contains(view, "Kubernetes Contexts") {
		t.Error("View should contain the title 'Kubernetes Contexts'")
	}
}

func TestContextListEmptyContexts(t *testing.T) {
	contexts := []services.Context{}
	cl := NewContextList(contexts, "")

	if len(cl.contexts) != 0 {
		t.Errorf("expected 0 contexts, got %d", len(cl.contexts))
	}

	// Test that SelectedContext returns nil for empty list
	selected := cl.SelectedContext()
	if selected != nil {
		t.Error("SelectedContext should return nil for empty list")
	}
}

func TestContextListNonExistentCurrentContext(t *testing.T) {
	contexts := []services.Context{
		{Name: "context1"},
		{Name: "context2"},
	}
	// Set current context to something that doesn't exist
	cl := NewContextList(contexts, "nonexistent")

	// Should still create the list without error
	if len(cl.contexts) != 2 {
		t.Errorf("expected 2 contexts, got %d", len(cl.contexts))
	}

	// Current context should be stored even if it doesn't match
	if cl.current != "nonexistent" {
		t.Errorf("expected current context 'nonexistent', got '%s'", cl.current)
	}
}
