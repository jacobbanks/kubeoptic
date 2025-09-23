package keys

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
)

func TestDefaultKeyMap(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test that all key bindings are properly defined
	testCases := []struct {
		name    string
		binding key.Binding
		keys    []string
	}{
		{"Up", keyMap.Up, []string{"up", "k"}},
		{"Down", keyMap.Down, []string{"down", "j"}},
		{"Left", keyMap.Left, []string{"left", "h"}},
		{"Right", keyMap.Right, []string{"right", "l"}},
		{"Enter", keyMap.Enter, []string{"enter"}},
		{"Back", keyMap.Back, []string{"esc", "backspace"}},
		{"Search", keyMap.Search, []string{"/"}},
		{"Follow", keyMap.Follow, []string{"f"}},
		{"Quit", keyMap.Quit, []string{"q", "ctrl+c"}},
		{"Help", keyMap.Help, []string{"?", "F1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.binding.Enabled() {
				t.Errorf("Key binding %s should be enabled", tc.name)
			}

			bindingKeys := tc.binding.Keys()
			if len(bindingKeys) != len(tc.keys) {
				t.Errorf("Key binding %s should have %d keys, got %d", tc.name, len(tc.keys), len(bindingKeys))
				return
			}

			for i, expectedKey := range tc.keys {
				if bindingKeys[i] != expectedKey {
					t.Errorf("Key binding %s key %d should be %s, got %s", tc.name, i, expectedKey, bindingKeys[i])
				}
			}
		})
	}
}

func TestGetViewKeyMap(t *testing.T) {
	keyMap := DefaultKeyMap()

	testCases := []struct {
		view         models.ViewType
		viewName     string
		expectedKeys int // minimum number of keys expected
	}{
		{models.ContextView, "ContextView", 5},
		{models.NamespaceView, "NamespaceView", 6},
		{models.PodView, "PodView", 8},
		{models.LogView, "LogView", 10},
	}

	for _, tc := range testCases {
		t.Run(tc.viewName, func(t *testing.T) {
			viewKeyMap := GetViewKeyMap(tc.view, keyMap)

			if len(viewKeyMap.View) < tc.expectedKeys {
				t.Errorf("View %v should have at least %d keys, got %d", tc.viewName, tc.expectedKeys, len(viewKeyMap.View))
			}

			if len(viewKeyMap.Help) == 0 {
				t.Errorf("View %v should have help keys defined", tc.viewName)
			}

			// Test that global keys are included
			if len(viewKeyMap.Global.Help.Keys()) == 0 {
				t.Error("Global help key should be defined")
			}
		})
	}
}

func TestHandleGlobalKeys(t *testing.T) {
	keyMap := DefaultKeyMap()

	testCases := []struct {
		name         string
		keyMsg       tea.KeyMsg
		shouldHandle bool
		expectQuit   bool
	}{
		{
			name:         "Quit with q",
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			shouldHandle: true,
			expectQuit:   true,
		},
		{
			name:         "Quit with ctrl+c",
			keyMsg:       tea.KeyMsg{Type: tea.KeyCtrlC},
			shouldHandle: true,
			expectQuit:   true,
		},
		{
			name:         "Help with ?",
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			shouldHandle: true,
			expectQuit:   false,
		},
		{
			name:         "Regular key",
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			shouldHandle: false,
			expectQuit:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handled, cmd := HandleGlobalKeys(tc.keyMsg, keyMap)

			if handled != tc.shouldHandle {
				t.Errorf("Expected handled=%v, got %v", tc.shouldHandle, handled)
			}

			if tc.expectQuit {
				if cmd == nil {
					t.Error("Expected quit command, got nil")
				}
				// Note: Can't easily test tea.Quit command directly,
				// just verify we got a command for quit actions
			} else if tc.shouldHandle {
				if cmd == nil {
					t.Error("Expected command, got nil")
				} else {
					// Execute the command and check message type
					msg := cmd()
					switch msg.(type) {
					case ShowHelpMsg:
						// Expected for help key
					default:
						t.Errorf("Unexpected message type: %T", msg)
					}
				}
			}
		})
	}
}

func TestIsNavigationKey(t *testing.T) {
	keyMap := DefaultKeyMap()

	testCases := []struct {
		name     string
		keyMsg   tea.KeyMsg
		expected bool
	}{
		{
			name:     "Up key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyUp},
			expected: true,
		},
		{
			name:     "Down key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyDown},
			expected: true,
		},
		{
			name:     "Left key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyLeft},
			expected: true,
		},
		{
			name:     "Right key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRight},
			expected: true,
		},
		{
			name:     "Tab key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyTab},
			expected: true,
		},
		{
			name:     "Enter key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyEnter},
			expected: false,
		},
		{
			name:     "Regular character",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNavigationKey(tc.keyMsg, keyMap)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestIsActionKey(t *testing.T) {
	keyMap := DefaultKeyMap()

	testCases := []struct {
		name     string
		keyMsg   tea.KeyMsg
		expected bool
	}{
		{
			name:     "Enter key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyEnter},
			expected: true,
		},
		{
			name:     "Escape key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyEsc},
			expected: true,
		},
		{
			name:     "Search key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			expected: true,
		},
		{
			name:     "Up key",
			keyMsg:   tea.KeyMsg{Type: tea.KeyUp},
			expected: false,
		},
		{
			name:     "Regular character",
			keyMsg:   tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsActionKey(tc.keyMsg, keyMap)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetKeyHelp(t *testing.T) {
	keyMap := DefaultKeyMap()

	testCases := []struct {
		view     models.ViewType
		viewName string
	}{
		{models.ContextView, "ContextView"},
		{models.NamespaceView, "NamespaceView"},
		{models.PodView, "PodView"},
		{models.LogView, "LogView"},
	}

	for _, tc := range testCases {
		t.Run(tc.viewName, func(t *testing.T) {
			help := GetKeyHelp(tc.view, keyMap)

			if len(help) == 0 {
				t.Errorf("View %v should have help sections", tc.viewName)
			}

			// Check that each section has key bindings
			for i, section := range help {
				if len(section) == 0 {
					t.Errorf("Help section %d for view %v should not be empty", i, tc.viewName)
				}

				// Check that each key binding in the section is enabled
				for j, binding := range section {
					if !binding.Enabled() {
						t.Errorf("Key binding %d in section %d for view %v should be enabled", j, i, tc.viewName)
					}
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkDefaultKeyMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultKeyMap()
	}
}

func BenchmarkGetViewKeyMap(b *testing.B) {
	keyMap := DefaultKeyMap()
	for i := 0; i < b.N; i++ {
		_ = GetViewKeyMap(models.PodView, keyMap)
	}
}

func BenchmarkHandleGlobalKeys(b *testing.B) {
	keyMap := DefaultKeyMap()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	for i := 0; i < b.N; i++ {
		_, _ = HandleGlobalKeys(msg, keyMap)
	}
}
