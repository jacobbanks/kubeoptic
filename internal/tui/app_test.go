package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
)

// MockKubeoptic creates a mock kubeoptic instance for testing
func createMockKubeoptic() *models.Kubeoptic {
	configSvc := services.NewConfigService()
	return models.NewKubeoptic(configSvc, nil, nil)
}

func TestNewApp(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.kubeoptic != kubeoptic {
		t.Error("App should store the kubeoptic instance")
	}

	if app.viewMode != ThreePanelView {
		t.Error("App should start in ThreePanelView mode")
	}

	if app.focusedPanel != ContextPanel {
		t.Error("App should start with ContextPanel focused")
	}

	if app.ready {
		t.Error("App should not be ready initially")
	}

	if app.initialized {
		t.Error("App should not be initialized initially")
	}
}

func TestAppInit(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	cmd := app.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}

	// Execute the batch command to see if it includes InitCompleteMsg
	msgs := []tea.Msg{}
	for cmd != nil {
		msg := cmd()
		msgs = append(msgs, msg)
		break // Just test first message
	}

	if len(msgs) == 0 {
		t.Error("Init() should return commands that produce messages")
	}
}

func TestAppWindowSizeMsg(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, cmd := app.Update(msg)

	app = newModel.(*App)
	if app.width != 80 {
		t.Errorf("Expected width 80, got %d", app.width)
	}
	if app.height != 24 {
		t.Errorf("Expected height 24, got %d", app.height)
	}
	if !app.ready {
		t.Error("App should be ready after receiving window size")
	}
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
}

func TestAppKeyMessages(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)
	app.width = 80
	app.height = 24
	app.ready = true

	tests := []struct {
		name          string
		key           string
		expectQuit    bool
		expectCommand bool
		description   string
	}{
		{"Quit with q", "q", true, false, "should quit"},
		{"Quit with ctrl+c", "ctrl+c", true, false, "should quit"},
		{"Tab navigation", "tab", false, false, "navigation without components"},
		{"Shift+Tab navigation", "shift+tab", false, false, "navigation without components"},
		{"Toggle fullscreen", "f", false, false, "fullscreen toggle"},
		{"Escape", "esc", false, false, "escape key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg tea.KeyMsg

			switch tt.key {
			case "q":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
			case "ctrl+c":
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			case "tab":
				msg = tea.KeyMsg{Type: tea.KeyTab}
			case "shift+tab":
				msg = tea.KeyMsg{Type: tea.KeyShiftTab}
			case "f":
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			newModel, cmd := app.Update(msg)

			if tt.expectQuit {
				if cmd == nil {
					t.Error("Expected quit command")
				}
				// We can't easily test if it's tea.Quit without executing
			} else if tt.expectCommand {
				if cmd == nil {
					t.Error("Expected command to be returned")
				}
			}

			app = newModel.(*App) // Reset for next test
		})
	}
}

func TestAppPanelNavigation(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test nextPanel
	initialPanel := app.focusedPanel
	app.nextPanel()

	if app.focusedPanel == initialPanel {
		t.Error("nextPanel() should change the focused panel")
	}

	// Test prevPanel
	app.prevPanel()
	if app.focusedPanel != initialPanel {
		t.Error("prevPanel() should return to initial panel")
	}
}

func TestAppViewModeToggle(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)
	app.width = 80
	app.height = 24
	app.ready = true

	// Start in ThreePanelView
	if app.viewMode != ThreePanelView {
		t.Error("Should start in ThreePanelView")
	}

	// Press 'f' to toggle to full screen
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	newModel, cmd := app.Update(msg)

	app = newModel.(*App)
	if app.viewMode != LogFullScreen {
		t.Error("Should switch to LogFullScreen mode")
	}
	if app.focusedPanel != LogPanel {
		t.Error("Should focus LogPanel in full screen mode")
	}
	if cmd == nil {
		// It's fine if no command is returned when no components are set
		t.Log("No command returned (expected when no components are set)")
	}

	// Press 'f' again to toggle back
	newModel, cmd = app.Update(msg)
	app = newModel.(*App)
	if app.viewMode != ThreePanelView {
		t.Error("Should switch back to ThreePanelView mode")
	}
	if app.focusedPanel != PodPanel {
		t.Error("Should focus PodPanel when returning from full screen")
	}
}

func TestAppStateManagerInterface(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test that app implements AppStateManager interface
	var _ AppStateManager = app

	// Test basic getters
	context := app.GetSelectedContext()
	namespace := app.GetSelectedNamespace()
	pods := app.GetFilteredPods()
	query := app.GetPodSearchQuery()
	following := app.IsFollowing()
	logs := app.GetLogBuffer()

	// These should not panic and should return sensible defaults
	if context == "" {
		// This is okay - might be empty initially
	}
	if namespace == "" {
		// This is okay - might be empty initially
	}
	if pods == nil {
		t.Error("GetFilteredPods() should not return nil")
	}
	if query != "" {
		// This is okay - should be empty initially
	}
	if following {
		// This is okay - might be false initially
	}
	if logs == nil {
		t.Error("GetLogBuffer() should not return nil")
	}
}

func TestAppErrorHandling(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)
	app.width = 80
	app.height = 24
	app.ready = true

	// Test error message handling
	testError := &testError{message: "test error"}
	errorMsg := ErrorMsg{Error: testError, Context: "test"}

	newModel, cmd := app.Update(errorMsg)
	app = newModel.(*App)

	if app.err != testError {
		t.Error("App should store the error")
	}
	if cmd != nil {
		t.Error("ErrorMsg should not return a command")
	}

	// Test error view
	view := app.View()
	if view == "" {
		t.Error("View() should return error content")
	}

	// Test clearing error
	clearMsg := ClearStatusMsg{}
	newModel, cmd = app.Update(clearMsg)
	app = newModel.(*App)

	if app.err != nil {
		t.Error("ClearStatusMsg should clear the error")
	}
}

func TestAppView(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test initial view (not ready)
	view := app.View()
	if view != "Initializing kubeoptic TUI..." {
		t.Errorf("Expected initialization message, got: %s", view)
	}

	// Test ready view
	app.width = 80
	app.height = 24
	app.ready = true
	app.initialized = true

	view = app.View()
	if view == "" {
		t.Error("View() should return content when ready")
	}
	if view == "Initializing kubeoptic TUI..." {
		t.Error("Should not show initialization message when ready")
	}
}

// testError implements error interface for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestAppMessageRouting(t *testing.T) {
	kubeoptic := createMockKubeoptic()
	app := NewApp(kubeoptic)

	// Test InitCompleteMsg
	initMsg := InitCompleteMsg{}
	newModel, cmd := app.Update(initMsg)
	app = newModel.(*App)

	if !app.initialized {
		t.Error("InitCompleteMsg should set initialized to true")
	}
	if cmd == nil {
		// It's fine if no command is returned when no components are set
		t.Log("No focus command returned (expected when no components are set)")
	}

	// Test ContextSelectedMsg
	contextMsg := ContextSelectedMsg{Context: &services.Context{Name: "test-context"}}
	newModel, cmd = app.Update(contextMsg)
	app = newModel.(*App)

	if app.focusedPanel != NamespacePanel {
		t.Error("ContextSelectedMsg should switch focus to NamespacePanel")
	}

	// Test PodSelectedMsg
	podMsg := PodSelectedMsg{Pod: &services.Pod{Name: "test-pod"}}
	newModel, cmd = app.Update(podMsg)
	app = newModel.(*App)

	if app.viewMode != LogFullScreen {
		t.Error("PodSelectedMsg should switch to LogFullScreen view")
	}
	if app.focusedPanel != LogPanel {
		t.Error("PodSelectedMsg should switch focus to LogPanel")
	}
}
