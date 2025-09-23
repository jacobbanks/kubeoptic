package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"kubeoptic/internal/tui/styles"
)

// TestStatusBarIntegration_WindowResize tests that the status bar properly handles window resize messages
func TestStatusBarIntegration_WindowResize(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Send window resize message
	msg := tea.WindowSizeMsg{Width: 120, Height: 30}
	updatedStatusBar, cmd := statusBar.Update(msg)

	if cmd != nil {
		t.Error("Status bar should not return commands on window resize")
	}

	if updatedStatusBar.width != 120 {
		t.Errorf("Expected width 120 after resize, got %d", updatedStatusBar.width)
	}

	// Test that the view respects the new width
	view := updatedStatusBar.View()
	renderedWidth := lipgloss.Width(view)
	if renderedWidth != 120 {
		t.Errorf("Expected rendered width 120, got %d", renderedWidth)
	}
}

// TestStatusBarIntegration_ConnectionStatusUpdate tests dynamic connection status updates
func TestStatusBarIntegration_ConnectionStatusUpdate(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)
	statusBar.SetSize(100, 1)

	// Test different connection states
	states := []struct {
		status   string
		expected string
	}{
		{"Connected", "Connected"},
		{"Connecting", "Connecting"},
		{"Disconnected", "Disconnected"},
		{"Error", "Error"},
		{"Reconnecting", "Reconnecting"},
	}

	for _, state := range states {
		statusBar.SetConnectionStatus(state.status)
		view := statusBar.View()

		if !strings.Contains(view, state.expected) {
			t.Errorf("Status bar should contain '%s' when connection status is set to '%s'", state.expected, state.status)
		}

		// Verify the status bar still maintains proper width
		if lipgloss.Width(view) != 100 {
			t.Errorf("Status bar should maintain width 100, got %d", lipgloss.Width(view))
		}
	}
}

// TestStatusBarIntegration_ResponsiveDesign tests responsive behavior across different screen sizes
func TestStatusBarIntegration_ResponsiveDesign(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test extreme cases that might cause layout issues
	testCases := []struct {
		name  string
		width int
	}{
		{"Very narrow", 20},
		{"Narrow", 40},
		{"Medium", 80},
		{"Wide", 120},
		{"Very wide", 200},
		{"Ultra wide", 300},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statusBar.SetSize(tc.width, 1)
			view := statusBar.View()

			renderedWidth := lipgloss.Width(view)
			if renderedWidth != tc.width {
				t.Errorf("Width %d: expected rendered width %d, got %d", tc.width, tc.width, renderedWidth)
			}

			// For very narrow widths, ensure we don't panic or render invalid content
			if tc.width < 30 && view == "" {
				// Empty view is acceptable for very narrow widths
				return
			}

			// For reasonable widths, ensure we have some content
			if tc.width >= 40 && len(strings.TrimSpace(view)) == 0 {
				t.Errorf("Width %d: status bar should have some content", tc.width)
			}
		})
	}
}

// TestStatusBarIntegration_StyleConsistency tests that styling is applied consistently
func TestStatusBarIntegration_StyleConsistency(t *testing.T) {
	themes := []styles.Theme{
		styles.DefaultTheme(),
		styles.DarkTheme(),
	}

	for i, theme := range themes {
		testName := "DefaultTheme"
		if i == 1 {
			testName = "DarkTheme"
		}

		t.Run(testName, func(t *testing.T) {
			statusBar := NewStatusBar(theme, nil)
			statusBar.SetSize(100, 1)

			// Test with different connection statuses to ensure styling is applied
			statuses := []string{"Connected", "Connecting", "Disconnected", "Error"}

			for _, status := range statuses {
				statusBar.SetConnectionStatus(status)
				view := statusBar.View()

				// Ensure the view is not empty and has proper width
				if len(strings.TrimSpace(view)) == 0 {
					t.Errorf("Theme test failed: status bar view should not be empty for status '%s'", status)
				}

				if lipgloss.Width(view) != 100 {
					t.Errorf("Theme test failed: expected width 100, got %d for status '%s'", lipgloss.Width(view), status)
				}
			}
		})
	}
}

// TestStatusBarIntegration_EdgeCases tests edge cases and boundary conditions
func TestStatusBarIntegration_EdgeCases(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	t.Run("Zero width", func(t *testing.T) {
		statusBar.SetSize(0, 1)
		view := statusBar.View()
		if view != "" {
			t.Error("Status bar with zero width should return empty string")
		}
	})

	t.Run("Negative width", func(t *testing.T) {
		statusBar.SetSize(-10, 1)
		view := statusBar.View()
		if view != "" {
			t.Error("Status bar with negative width should return empty string")
		}
	})

	t.Run("Empty connection status", func(t *testing.T) {
		statusBar.SetSize(100, 1)
		statusBar.SetConnectionStatus("")
		view := statusBar.View()

		// Should still render something (just no connection status)
		if lipgloss.Width(view) != 100 {
			t.Errorf("Expected width 100 even with empty connection status, got %d", lipgloss.Width(view))
		}
	})

	t.Run("Very long connection status", func(t *testing.T) {
		statusBar.SetSize(50, 1)
		longStatus := "Very long connection status that exceeds the available width significantly"
		statusBar.SetConnectionStatus(longStatus)
		view := statusBar.View()

		// Should still maintain the set width
		if lipgloss.Width(view) != 50 {
			t.Errorf("Expected width 50 even with very long connection status, got %d", lipgloss.Width(view))
		}
	})
}

// TestStatusBarIntegration_TeaModelInterface tests Bubble Tea integration
func TestStatusBarIntegration_TeaModelInterface(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test that the component properly implements tea.Model interface behavior
	t.Run("Update returns proper types", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		updatedStatusBar, cmd := statusBar.Update(msg)

		if updatedStatusBar == nil {
			t.Error("Update should return a non-nil StatusBar")
		}

		// For window resize, cmd should be nil
		if cmd != nil {
			t.Error("Window resize should not return a command")
		}
	})

	t.Run("View returns string", func(t *testing.T) {
		statusBar.SetSize(80, 1)
		view := statusBar.View()

		// View should return a string (may be empty for zero width, but still a string)
		if view == "" && statusBar.width > 0 {
			t.Error("View should return non-empty string for non-zero width")
		}
	})

	t.Run("Unknown message types", func(t *testing.T) {
		// Test with an unknown message type
		type unknownMsg struct{}
		updatedStatusBar, cmd := statusBar.Update(unknownMsg{})

		if updatedStatusBar == nil {
			t.Error("Update should return StatusBar even for unknown messages")
		}

		if cmd != nil {
			t.Error("Unknown messages should not return commands")
		}
	})
}
