package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"kubeoptic/internal/tui/styles"
)

func TestNewStatusBar(t *testing.T) {
	theme := styles.DefaultTheme()

	statusBar := NewStatusBar(theme, nil)

	if statusBar == nil {
		t.Fatal("NewStatusBar should return a non-nil StatusBar")
	}

	if statusBar.theme != theme {
		t.Error("StatusBar should store the provided theme")
	}

	if statusBar.height != 1 {
		t.Error("StatusBar should have default height of 1")
	}

	if statusBar.connectionStatus != "Connected" {
		t.Error("StatusBar should have default connection status of 'Connected'")
	}
}

func TestStatusBarSetSize(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	statusBar.SetSize(100, 2)

	if statusBar.width != 100 {
		t.Errorf("Expected width 100, got %d", statusBar.width)
	}

	if statusBar.height != 2 {
		t.Errorf("Expected height 2, got %d", statusBar.height)
	}

	// Test that height 0 doesn't change height
	statusBar.SetSize(120, 0)
	if statusBar.width != 120 {
		t.Errorf("Expected width 120, got %d", statusBar.width)
	}
	if statusBar.height != 2 {
		t.Errorf("Expected height to remain 2, got %d", statusBar.height)
	}
}

func TestStatusBarSetConnectionStatus(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	statusBar.SetConnectionStatus("Disconnected")

	if statusBar.connectionStatus != "Disconnected" {
		t.Errorf("Expected connection status 'Disconnected', got '%s'", statusBar.connectionStatus)
	}
}

func TestStatusBarViewEmptyWidth(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	view := statusBar.View()

	if view != "" {
		t.Error("StatusBar with zero width should return empty string")
	}
}

func TestStatusBarViewWithWidth(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)
	statusBar.SetSize(80, 1)

	view := statusBar.View()

	if view == "" {
		t.Error("StatusBar with width should return non-empty string")
	}

	// Check that the rendered width matches the set width
	renderedWidth := lipgloss.Width(view)
	if renderedWidth != 80 {
		t.Errorf("Expected rendered width 80, got %d", renderedWidth)
	}
}

func TestStatusBarBuildLeftSection(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test with nil kubeoptic
	left := statusBar.buildLeftSection()
	if left != "" {
		t.Error("buildLeftSection should return empty string when kubeoptic is nil")
	}
}

func TestStatusBarBuildRightSection(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test default connection status
	right := statusBar.buildRightSection()
	if !strings.Contains(right, "Connected") {
		t.Error("Right section should contain connection status")
	}

	// Test custom connection status
	statusBar.SetConnectionStatus("Disconnected")
	right = statusBar.buildRightSection()
	if !strings.Contains(right, "Disconnected") {
		t.Error("Right section should contain updated connection status")
	}

	// Test empty connection status
	statusBar.SetConnectionStatus("")
	right = statusBar.buildRightSection()
	if right != "" {
		t.Error("Right section should be empty when connection status is empty")
	}
}

func TestStatusBarBuildCenterSection(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test with zero available width
	center := statusBar.buildCenterSection(0)
	if center != "" {
		t.Error("Center section should be empty when available width is 0")
	}

	// Test with available width but nil kubeoptic
	center = statusBar.buildCenterSection(50)
	if lipgloss.Width(center) != 50 {
		t.Error("Center section should fill available width")
	}
}

func TestStatusBarTruncateCenter(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test with very small width
	result := statusBar.truncateCenter("very long text that should be truncated", 5)
	if lipgloss.Width(result) > 5 {
		t.Error("Truncated text should not exceed max width")
	}

	// Test with pod information preservation
	podText := "pod:my-very-long-pod-name │ 5 of 10 pods"
	result = statusBar.truncateCenter(podText, 20)
	if !strings.Contains(result, "pod:") {
		t.Error("Truncation should preserve pod information when possible")
	}
}

func TestStatusBarResponsiveLayout(t *testing.T) {
	statusBar := NewStatusBar(styles.DefaultTheme(), nil)

	// Test with different widths
	widths := []int{40, 80, 120, 200}

	for _, width := range widths {
		statusBar.SetSize(width, 1)
		view := statusBar.View()

		renderedWidth := lipgloss.Width(view)
		if renderedWidth != width {
			t.Errorf("Status bar should adapt to width %d, got %d", width, renderedWidth)
		}

		// Ensure content is present for reasonable widths
		if width >= 80 && view == "" {
			t.Errorf("Status bar should have content for width %d", width)
		}
	}
}

func TestStatusBarMaxFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 3, 5},
		{3, 5, 5},
		{5, 5, 5},
		{0, 1, 1},
		{-1, 0, 0},
	}

	for _, test := range tests {
		result := max(test.a, test.b)
		if result != test.expected {
			t.Errorf("max(%d, %d) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}
