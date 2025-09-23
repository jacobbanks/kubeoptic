package navigation

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/tui/keys"
)

func TestNewNavigationState(t *testing.T) {
	nav := NewNavigationState()

	if nav.currentView != models.ContextView {
		t.Errorf("Expected initial view to be ContextView, got %v", nav.currentView)
	}

	if nav.focusedPanel != FocusContext {
		t.Errorf("Expected initial focus to be FocusContext, got %v", nav.focusedPanel)
	}

	if nav.helpVisible {
		t.Error("Expected help to be initially hidden")
	}

	if nav.searchMode {
		t.Error("Expected search mode to be initially disabled")
	}

	if len(nav.navigationHistory) != 1 {
		t.Errorf("Expected navigation history to have 1 item, got %d", len(nav.navigationHistory))
	}
}

func TestNavigateToView(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		name          string
		targetView    models.ViewType
		expectedFocus FocusState
		expectCmd     bool
	}{
		{"Navigate to Namespace", models.NamespaceView, FocusNamespace, true},
		{"Navigate to Pod", models.PodView, FocusPod, true},
		{"Navigate to Log", models.LogView, FocusLog, true},
		{"Navigate to same view", models.ContextView, FocusContext, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset to context view for each test
			nav.currentView = models.ContextView
			nav.focusedPanel = FocusContext

			cmd := nav.NavigateToView(tc.targetView)

			if nav.currentView != tc.targetView {
				t.Errorf("Expected current view to be %v, got %v", tc.targetView, nav.currentView)
			}

			if nav.focusedPanel != tc.expectedFocus {
				t.Errorf("Expected focused panel to be %v, got %v", tc.expectedFocus, nav.focusedPanel)
			}

			if tc.expectCmd && cmd == nil {
				t.Error("Expected command to be returned")
			} else if !tc.expectCmd && cmd != nil {
				t.Error("Expected no command for same view navigation")
			}

			// Test search mode is cleared
			if nav.searchMode {
				t.Error("Search mode should be cleared when changing views")
			}
		})
	}
}

func TestNavigateBack(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		name         string
		currentView  models.ViewType
		expectedView models.ViewType
		expectCmd    bool
	}{
		{"Back from Log", models.LogView, models.PodView, true},
		{"Back from Pod", models.PodView, models.NamespaceView, true},
		{"Back from Namespace", models.NamespaceView, models.ContextView, true},
		{"Back from Context", models.ContextView, models.ContextView, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nav.currentView = tc.currentView

			cmd := nav.NavigateBack()

			if tc.expectCmd {
				if cmd == nil {
					t.Error("Expected command to be returned")
				}
				// Execute the command to trigger the navigation
				if cmd != nil {
					cmd()
				}
				if nav.currentView != tc.expectedView {
					t.Errorf("Expected view to be %v, got %v", tc.expectedView, nav.currentView)
				}
			} else {
				if cmd != nil {
					t.Error("Expected no command for context view back navigation")
				}
			}
		})
	}
}

func TestNextPanel(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		name          string
		view          models.ViewType
		currentFocus  FocusState
		expectedFocus FocusState
		expectCmd     bool
	}{
		{"Namespace view: Context to Namespace", models.NamespaceView, FocusContext, FocusNamespace, true},
		{"Namespace view: Namespace to Context", models.NamespaceView, FocusNamespace, FocusContext, true},
		{"Pod view: Context to Namespace", models.PodView, FocusContext, FocusNamespace, true},
		{"Pod view: Namespace to Pod", models.PodView, FocusNamespace, FocusPod, true},
		{"Pod view: Pod to Context", models.PodView, FocusPod, FocusContext, true},
		{"Context view: No panel change", models.ContextView, FocusContext, FocusContext, false},
		{"Log view: No panel change", models.LogView, FocusLog, FocusLog, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nav.currentView = tc.view
			nav.focusedPanel = tc.currentFocus

			cmd := nav.NextPanel()

			if nav.focusedPanel != tc.expectedFocus {
				t.Errorf("Expected focus to be %v, got %v", tc.expectedFocus, nav.focusedPanel)
			}

			if tc.expectCmd && cmd == nil {
				t.Error("Expected command to be returned")
			} else if !tc.expectCmd && cmd != nil {
				t.Error("Expected no command for single-panel views")
			}
		})
	}
}

func TestPrevPanel(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		name          string
		view          models.ViewType
		currentFocus  FocusState
		expectedFocus FocusState
		expectCmd     bool
	}{
		{"Namespace view: Context to Namespace", models.NamespaceView, FocusContext, FocusNamespace, true},
		{"Namespace view: Namespace to Context", models.NamespaceView, FocusNamespace, FocusContext, true},
		{"Pod view: Context to Pod", models.PodView, FocusContext, FocusPod, true},
		{"Pod view: Namespace to Context", models.PodView, FocusNamespace, FocusContext, true},
		{"Pod view: Pod to Namespace", models.PodView, FocusPod, FocusNamespace, true},
		{"Context view: No panel change", models.ContextView, FocusContext, FocusContext, false},
		{"Log view: No panel change", models.LogView, FocusLog, FocusLog, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nav.currentView = tc.view
			nav.focusedPanel = tc.currentFocus

			cmd := nav.PrevPanel()

			if nav.focusedPanel != tc.expectedFocus {
				t.Errorf("Expected focus to be %v, got %v", tc.expectedFocus, nav.focusedPanel)
			}

			if tc.expectCmd && cmd == nil {
				t.Error("Expected command to be returned")
			} else if !tc.expectCmd && cmd != nil {
				t.Error("Expected no command for single-panel views")
			}
		})
	}
}

func TestToggleHelp(t *testing.T) {
	nav := NewNavigationState()

	// Initially help should be hidden
	if nav.helpVisible {
		t.Error("Help should be initially hidden")
	}

	// Toggle to show help
	cmd := nav.ToggleHelp()
	if !nav.helpVisible {
		t.Error("Help should be visible after toggle")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	// Check the message type
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(keys.ShowHelpMsg); !ok {
			t.Errorf("Expected ShowHelpMsg, got %T", msg)
		}
	}

	// Toggle to hide help
	cmd = nav.ToggleHelp()
	if nav.helpVisible {
		t.Error("Help should be hidden after second toggle")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	// Check the message type
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(keys.HideHelpMsg); !ok {
			t.Errorf("Expected HideHelpMsg, got %T", msg)
		}
	}
}

func TestSearchMode(t *testing.T) {
	nav := NewNavigationState()

	// Test entering search mode
	cmd := nav.EnterSearchMode()
	if !nav.searchMode {
		t.Error("Should be in search mode after EnterSearchMode()")
	}
	if nav.focusedPanel != FocusSearch {
		t.Error("Focus should be on search when entering search mode")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	// Test exiting search mode
	nav.currentView = models.PodView // Set a view to test focus restoration
	cmd = nav.ExitSearchMode()
	if nav.searchMode {
		t.Error("Should not be in search mode after ExitSearchMode()")
	}
	if nav.focusedPanel != FocusPod {
		t.Error("Focus should be restored to pod panel after exiting search")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestCanNavigateForward(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		view     models.ViewType
		expected bool
	}{
		{models.ContextView, true},
		{models.NamespaceView, true},
		{models.PodView, true},
		{models.LogView, false},
	}

	for _, tc := range testCases {
		t.Run("View_"+string(rune(tc.view)), func(t *testing.T) {
			nav.currentView = tc.view
			result := nav.CanNavigateForward()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for view %v", tc.expected, result, tc.view)
			}
		})
	}
}

func TestCanNavigateBack(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		view     models.ViewType
		expected bool
	}{
		{models.ContextView, false},
		{models.NamespaceView, true},
		{models.PodView, true},
		{models.LogView, true},
	}

	for _, tc := range testCases {
		t.Run("View_"+string(rune(tc.view)), func(t *testing.T) {
			nav.currentView = tc.view
			result := nav.CanNavigateBack()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for view %v", tc.expected, result, tc.view)
			}
		})
	}
}

func TestGetBreadcrumb(t *testing.T) {
	nav := NewNavigationState()

	testCases := []struct {
		view     models.ViewType
		expected string
	}{
		{models.ContextView, "Contexts"},
		{models.NamespaceView, "Contexts > Namespaces"},
		{models.PodView, "Contexts > Namespaces > Pods"},
		{models.LogView, "Contexts > Namespaces > Pods > Logs"},
	}

	for _, tc := range testCases {
		t.Run("View_"+string(rune(tc.view)), func(t *testing.T) {
			nav.currentView = tc.view
			result := nav.GetBreadcrumb()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q for view %v", tc.expected, result, tc.view)
			}
		})
	}
}

func TestHandleNavigation(t *testing.T) {
	nav := NewNavigationState()
	keyMap := keys.DefaultKeyMap()

	testCases := []struct {
		name      string
		keyMsg    tea.KeyMsg
		expectCmd bool
	}{
		{
			name:      "Back key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyEsc},
			expectCmd: false, // No command for back from context view
		},
		{
			name:      "Tab key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyTab},
			expectCmd: false, // No command for single-panel view
		},
		{
			name:      "Search key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			expectCmd: true,
		},
		{
			name:      "Help key",
			keyMsg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			expectCmd: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset navigation state
			nav.currentView = models.ContextView
			nav.focusedPanel = FocusContext
			nav.helpVisible = false
			nav.searchMode = false

			cmd := nav.HandleNavigation(tc.keyMsg, keyMap)

			if tc.expectCmd && cmd == nil {
				t.Error("Expected command to be returned")
			} else if !tc.expectCmd && cmd != nil {
				t.Error("Expected no command to be returned")
			}
		})
	}
}

func TestIsValidTransition(t *testing.T) {
	testCases := []struct {
		from     models.ViewType
		to       models.ViewType
		expected bool
	}{
		{models.ContextView, models.NamespaceView, true},
		{models.NamespaceView, models.PodView, true},
		{models.PodView, models.LogView, true},
		{models.LogView, models.PodView, true},
		{models.PodView, models.NamespaceView, true},
		{models.NamespaceView, models.ContextView, true},
		{models.ContextView, models.LogView, false}, // Invalid direct transition
		{models.LogView, models.ContextView, false}, // Invalid direct transition
	}

	for _, tc := range testCases {
		t.Run("From_To", func(t *testing.T) {
			result := IsValidTransition(tc.from, tc.to)
			if result != tc.expected {
				t.Errorf("Expected %v for transition from %v to %v, got %v", tc.expected, tc.from, tc.to, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNavigateToView(b *testing.B) {
	nav := NewNavigationState()
	for i := 0; i < b.N; i++ {
		nav.NavigateToView(models.PodView)
		nav.NavigateToView(models.ContextView)
	}
}

func BenchmarkNextPanel(b *testing.B) {
	nav := NewNavigationState()
	nav.currentView = models.PodView
	for i := 0; i < b.N; i++ {
		nav.NextPanel()
	}
}

func BenchmarkHandleNavigation(b *testing.B) {
	nav := NewNavigationState()
	keyMap := keys.DefaultKeyMap()
	msg := tea.KeyMsg{Type: tea.KeyTab}

	for i := 0; i < b.N; i++ {
		nav.HandleNavigation(msg, keyMap)
	}
}
