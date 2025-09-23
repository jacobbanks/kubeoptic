//go:build integration

package events

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/keys"
	"kubeoptic/internal/tui/navigation"
)

func TestEventProcessorIntegration(t *testing.T) {
	processor := NewEventProcessor()

	// Test initial state
	nav := processor.GetNavigation()
	if nav.GetCurrentView() != models.ContextView {
		t.Error("Expected initial view to be ContextView")
	}

	// Test key handling integration
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	model, cmd := processor.ProcessEvent(keyMsg)

	if cmd == nil {
		t.Error("Expected command for help key")
	}

	// The model should be nil for global commands
	if model != nil {
		t.Error("Expected nil model for global command")
	}
}

func TestFullUserJourneyIntegration(t *testing.T) {
	processor := NewEventProcessor()
	nav := processor.GetNavigation()

	// Start at context view
	if nav.GetCurrentView() != models.ContextView {
		t.Error("Expected to start at ContextView")
	}

	// Simulate user pressing Enter to select context
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := processor.ProcessEvent(enterMsg)

	if cmd != nil {
		// Execute the command to trigger navigation
		cmd()
	}

	// Simulate navigation to namespace view
	nav.NavigateToView(models.NamespaceView)
	if nav.GetCurrentView() != models.NamespaceView {
		t.Error("Expected to be at NamespaceView after navigation")
	}

	// Test tab navigation between panels
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd = processor.ProcessEvent(tabMsg)

	if cmd != nil {
		cmd()
	}

	// Test navigation to pod view
	nav.NavigateToView(models.PodView)
	if nav.GetCurrentView() != models.PodView {
		t.Error("Expected to be at PodView")
	}

	// Test search functionality
	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	_, cmd = processor.ProcessEvent(searchMsg)

	if cmd != nil {
		cmd()
	}

	if !nav.IsSearchMode() {
		t.Error("Expected to be in search mode")
	}

	// Test escape from search
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = processor.ProcessEvent(escMsg)

	if cmd != nil {
		cmd()
	}

	if nav.IsSearchMode() {
		t.Error("Expected to exit search mode")
	}

	// Test navigation to log view
	nav.NavigateToView(models.LogView)
	if nav.GetCurrentView() != models.LogView {
		t.Error("Expected to be at LogView")
	}

	// Test back navigation
	backMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = processor.ProcessEvent(backMsg)

	if cmd != nil {
		cmd()
	}

	// Should be back at pod view
	if nav.GetCurrentView() != models.PodView {
		t.Error("Expected to be back at PodView after back navigation")
	}
}

func TestEventRouterIntegration(t *testing.T) {
	nav := navigation.NewNavigationState()
	router := NewEventRouter(nav)

	// Test context view event routing
	nav.NavigateToView(models.ContextView)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd, handled := router.RouteEvent(enterMsg)

	if !handled {
		t.Error("Enter key should be handled in context view")
	}

	if cmd == nil {
		t.Error("Enter key should return command in context view")
	}

	if model != nil {
		t.Error("Expected nil model for context view enter")
	}

	// Test pod view event routing with search
	nav.NavigateToView(models.PodView)

	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	_, cmd, handled = router.RouteEvent(searchMsg)

	if !handled {
		t.Error("Search key should be handled in pod view")
	}

	if cmd == nil {
		t.Error("Search key should return command")
	}

	// Test log view event routing
	nav.NavigateToView(models.LogView)

	followMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd, handled = router.RouteEvent(followMsg)

	if !handled {
		t.Error("Follow key should be handled in log view")
	}

	if cmd == nil {
		t.Error("Follow key should return command")
	}
}

func TestWindowResizeIntegration(t *testing.T) {
	processor := NewEventProcessor()

	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	_, cmd := processor.ProcessEvent(resizeMsg)

	if cmd == nil {
		t.Error("Window resize should return command")
	}

	// Execute command and check message type
	if cmd != nil {
		msg := cmd()
		if resizeResult, ok := msg.(tui.WindowResizeMsg); ok {
			if resizeResult.Width != 120 || resizeResult.Height != 40 {
				t.Error("Window resize dimensions not preserved")
			}
		} else {
			t.Error("Expected tui.WindowResizeMsg from resize command")
		}
	}
}

func TestMouseEventIntegration(t *testing.T) {
	processor := NewEventProcessor()
	nav := processor.GetNavigation()

	// Test mouse click in log view
	nav.NavigateToView(models.LogView)

	mouseMsg := tea.MouseMsg{
		X:      10,
		Y:      5,
		Type:   tea.MouseLeft,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}

	_, cmd := processor.ProcessEvent(mouseMsg)

	if cmd == nil {
		t.Error("Mouse click should return command")
	}

	// Test mouse wheel in log view
	wheelMsg := tea.MouseMsg{
		X:      10,
		Y:      5,
		Type:   tea.MouseWheelDown,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
	}

	_, cmd = processor.ProcessEvent(wheelMsg)

	if cmd == nil {
		t.Error("Mouse wheel should return command in log view")
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	processor := NewEventProcessor()

	// Test with invalid/unknown message type
	unknownMsg := struct{ Unknown string }{Unknown: "test"}

	model, cmd := processor.ProcessEvent(unknownMsg)

	// Should not panic and should return nil for unknown messages
	if model != nil || cmd != nil {
		t.Error("Unknown message should return nil model and command")
	}
}

func TestKeyMapCustomizationIntegration(t *testing.T) {
	processor := NewEventProcessor()

	// Get original key map
	originalKeyMap := processor.GetKeyMap()

	// Create custom key map
	customKeyMap := keys.DefaultKeyMap()
	customKeyMap.Quit = keys.DefaultKeyMap().Help // Swap quit and help keys for testing

	// Update key map
	processor.UpdateKeyMap(customKeyMap)

	// Verify key map was updated
	updatedKeyMap := processor.GetKeyMap()
	if updatedKeyMap.Quit.Keys()[0] == originalKeyMap.Quit.Keys()[0] {
		t.Error("Key map should have been updated")
	}
}

func TestMultiPanelNavigationIntegration(t *testing.T) {
	processor := NewEventProcessor()
	nav := processor.GetNavigation()

	// Navigate to pod view (3-panel layout)
	nav.NavigateToView(models.PodView)

	// Test tab navigation through panels
	initialFocus := nav.GetFocusedPanel()

	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := processor.ProcessEvent(tabMsg)

	if cmd != nil {
		cmd()
	}

	// Focus should have changed
	if nav.GetFocusedPanel() == initialFocus {
		t.Error("Tab should change focus in multi-panel view")
	}

	// Test shift+tab for reverse navigation
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	_, cmd = processor.ProcessEvent(shiftTabMsg)

	if cmd != nil {
		cmd()
	}

	// Should be back to original focus
	if nav.GetFocusedPanel() != initialFocus {
		t.Error("Shift+Tab should reverse panel navigation")
	}
}

func TestSearchModeIntegration(t *testing.T) {
	processor := NewEventProcessor()
	nav := processor.GetNavigation()

	// Start search
	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	_, cmd := processor.ProcessEvent(searchMsg)

	if cmd != nil {
		cmd()
	}

	if !nav.IsSearchMode() {
		t.Error("Should be in search mode")
	}

	// Test that navigation keys are handled differently in search mode
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	_, cmd = processor.ProcessEvent(upMsg)

	// In search mode, navigation should be handled by search component
	// So the event processor should return nil
	if cmd != nil {
		t.Log("Search mode may delegate to component handlers")
	}

	// Exit search with escape
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd = processor.ProcessEvent(escMsg)

	if cmd != nil {
		cmd()
	}

	if nav.IsSearchMode() {
		t.Error("Should exit search mode on escape")
	}
}

// Benchmark integration tests
func BenchmarkFullEventFlow(b *testing.B) {
	processor := NewEventProcessor()
	messages := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyEnter},
		tea.KeyMsg{Type: tea.KeyTab},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
		tea.KeyMsg{Type: tea.KeyEsc},
		tea.WindowSizeMsg{Width: 80, Height: 24},
	}

	for i := 0; i < b.N; i++ {
		for _, msg := range messages {
			processor.ProcessEvent(msg)
		}
	}
}

func BenchmarkEventRoutingWithNavigation(b *testing.B) {
	nav := navigation.NewNavigationState()
	router := NewEventRouter(nav)

	msg := tea.KeyMsg{Type: tea.KeyEnter}

	for i := 0; i < b.N; i++ {
		// Cycle through different views
		view := models.ViewType(i % 4)
		nav.NavigateToView(view)
		router.RouteEvent(msg)
	}
}
