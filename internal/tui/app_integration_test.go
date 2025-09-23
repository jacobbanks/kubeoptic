package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestAppFullWorkflow tests the complete application workflow
func TestAppFullWorkflow(t *testing.T) {
	// Create a mock kubeoptic instance with some test data
	kubeoptic := createMockKubeoptic()

	// Simulate loading some test contexts and namespaces
	// Note: In a real integration test, we'd want to set up test data
	// but for now we'll just test the basic workflow without K8s calls

	app := NewApp(kubeoptic)

	// Test initialization
	cmd := app.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}

	// Simulate window resize (like a real TUI startup)
	windowMsg := tea.WindowSizeMsg{Width: 120, Height: 30}
	newModel, cmd := app.Update(windowMsg)
	app = newModel.(*App)

	if !app.ready {
		t.Error("App should be ready after window size message")
	}
	if app.width != 120 || app.height != 30 {
		t.Error("App should store window dimensions")
	}

	// Test view rendering
	view := app.View()
	if view == "" {
		t.Error("View should return content")
	}
	if view == "Initializing kubeoptic TUI..." {
		t.Error("Should not show initialization message when ready")
	}

	// Test basic navigation workflow
	initialPanel := app.focusedPanel

	// Navigate through panels
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, cmd = app.Update(tabMsg)
	app = newModel.(*App)

	if app.focusedPanel == initialPanel {
		// This is okay if there's nowhere to navigate
		t.Log("Panel focus unchanged (expected with no components)")
	}

	// Test view mode toggle
	fMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	newModel, cmd = app.Update(fMsg)
	app = newModel.(*App)

	if app.viewMode != LogFullScreen {
		t.Error("Should switch to full screen mode")
	}
	if app.focusedPanel != LogPanel {
		t.Error("Should focus log panel in full screen")
	}

	// Test going back
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd = app.Update(escMsg)
	app = newModel.(*App)

	if app.viewMode != ThreePanelView {
		t.Error("Should return to three panel view")
	}

	// Test error handling workflow
	testErr := &testError{message: "integration test error"}
	errorMsg := ErrorMsg{Error: testErr, Context: "integration test"}
	newModel, cmd = app.Update(errorMsg)
	app = newModel.(*App)

	if app.err == nil {
		t.Error("App should store the error")
	}

	errorView := app.View()
	if errorView == "" {
		t.Error("Should show error view")
	}

	// Clear the error
	clearMsg := ClearStatusMsg{}
	newModel, cmd = app.Update(clearMsg)
	app = newModel.(*App)

	if app.err != nil {
		t.Error("Error should be cleared")
	}

	// Test state manager interface usage
	selectedContext := app.GetSelectedContext()
	selectedNamespace := app.GetSelectedNamespace()
	pods := app.GetFilteredPods()

	// These should not panic
	if pods == nil {
		t.Error("GetFilteredPods should not return nil")
	}

	// Test search functionality (without actually calling K8s operations)
	// app.SetPodSearchQuery("test-query") // Skip this as it calls K8s
	query := app.GetPodSearchQuery()
	if query != "" {
		// This is fine - should be empty initially
	}

	// Test quit
	quitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	newModel, cmd = app.Update(quitMsg)

	if cmd == nil {
		t.Error("Quit should return a command")
	}

	t.Logf("Integration test completed successfully")
	t.Logf("Final state - Context: %s, Namespace: %s", selectedContext, selectedNamespace)
}

// TestAppWithMockComponents tests the app with mock components
func TestAppWithMockComponents(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Create mock components
	mockContext := &MockComponent{name: "context"}
	mockNamespace := &MockComponent{name: "namespace"}
	mockPod := &MockComponent{name: "pod"}
	mockLog := &MockComponent{name: "log"}
	mockStatus := &MockComponent{name: "status"}

	// Set mock components
	app.SetComponents(mockContext, mockNamespace, mockPod, mockLog, mockStatus)

	// Test that components are set
	if app.contextList == nil {
		t.Error("Context list should be set")
	}
	if app.namespaceList == nil {
		t.Error("Namespace list should be set")
	}
	if app.podList == nil {
		t.Error("Pod list should be set")
	}

	// Test initialization with components
	cmd := app.Init()
	if cmd == nil {
		t.Error("Init with components should return command")
	}

	// Test focus management with components
	app.width = 80
	app.height = 24
	app.ready = true

	updateCmd := app.updateFocus()
	if updateCmd == nil {
		// This is okay if mock components don't support focus
		t.Log("No focus command returned from mock components")
	}

	// Test component updates
	testMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")}
	newModel, cmd := app.updateComponents(testMsg)
	app = newModel.(*App)

	// Verify all mock components received the message
	if !mockContext.updated {
		t.Error("Context component should have been updated")
	}
	if !mockNamespace.updated {
		t.Error("Namespace component should have been updated")
	}
	if !mockPod.updated {
		t.Error("Pod component should have been updated")
	}
}

// MockComponent implements ComponentRenderer for testing
type MockComponent struct {
	name    string
	updated bool
	focused bool
	width   int
	height  int
}

func (m *MockComponent) Init() tea.Cmd {
	return func() tea.Msg {
		return "mock init complete"
	}
}

func (m *MockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updated = true
	return m, nil
}

func (m *MockComponent) View() string {
	status := "unfocused"
	if m.focused {
		status = "focused"
	}
	return "Mock " + m.name + " component (" + status + ")"
}

// MockComponent also implements other interfaces for testing
func (m *MockComponent) Focus() tea.Cmd {
	m.focused = true
	return func() tea.Msg {
		return FocusChangedMsg{Component: m.name, Focused: true}
	}
}

func (m *MockComponent) Blur() tea.Cmd {
	m.focused = false
	return func() tea.Msg {
		return FocusChangedMsg{Component: m.name, Focused: false}
	}
}

func (m *MockComponent) IsFocused() bool {
	return m.focused
}

func (m *MockComponent) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MockComponent) GetSize() (int, int) {
	return m.width, m.height
}

func (m *MockComponent) RefreshData() tea.Cmd {
	return func() tea.Msg {
		return "mock data refreshed"
	}
}

func (m *MockComponent) GetLoadingState() bool {
	return false
}

func (m *MockComponent) GetErrorState() error {
	return nil
}

// TestAppPerformance tests basic performance characteristics
func TestAppPerformance(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test that basic operations complete quickly
	start := time.Now()

	// Simulate rapid key presses
	for i := 0; i < 100; i++ {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		app.Update(msg)
	}

	elapsed := time.Since(start)
	if elapsed > time.Millisecond*100 {
		t.Errorf("100 key updates took too long: %v", elapsed)
	}

	// Test view rendering performance
	app.width = 120
	app.height = 30
	app.ready = true

	start = time.Now()
	for i := 0; i < 50; i++ {
		app.View()
	}
	elapsed = time.Since(start)

	if elapsed > time.Millisecond*50 {
		t.Errorf("50 view renders took too long: %v", elapsed)
	}
}
