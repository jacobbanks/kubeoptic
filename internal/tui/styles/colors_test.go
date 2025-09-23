package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	// Test that all theme colors are set
	if theme.Primary == "" {
		t.Error("Primary color should not be empty")
	}
	if theme.Secondary == "" {
		t.Error("Secondary color should not be empty")
	}
	if theme.Background == "" {
		t.Error("Background color should not be empty")
	}
	if theme.Foreground == "" {
		t.Error("Foreground color should not be empty")
	}
	if theme.Border == "" {
		t.Error("Border color should not be empty")
	}
	if theme.Highlight == "" {
		t.Error("Highlight color should not be empty")
	}
	if theme.Error == "" {
		t.Error("Error color should not be empty")
	}
	if theme.Warning == "" {
		t.Error("Warning color should not be empty")
	}
	if theme.Success == "" {
		t.Error("Success color should not be empty")
	}
	if theme.Info == "" {
		t.Error("Info color should not be empty")
	}
}

func TestDarkTheme(t *testing.T) {
	theme := DarkTheme()

	// Test that dark theme has different values than default
	defaultTheme := DefaultTheme()
	if theme.Primary == defaultTheme.Primary {
		t.Error("Dark theme primary should differ from default theme")
	}

	// Test that all colors are valid
	if theme.Background == "" {
		t.Error("Dark theme background should not be empty")
	}
}

func TestGetStatusColor(t *testing.T) {
	tests := []struct {
		status   string
		expected lipgloss.Color
	}{
		{"Running", StatusRunning},
		{"Pending", StatusPending},
		{"Failed", StatusFailed},
		{"Succeeded", StatusSucceeded},
		{"Terminated", StatusTerminated},
		{"Unknown", StatusUnknown},
		{"InvalidStatus", StatusUnknown},
		{"", StatusUnknown},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			result := GetStatusColor(test.status)
			if result != test.expected {
				t.Errorf("GetStatusColor(%q) = %v, want %v", test.status, result, test.expected)
			}
		})
	}
}

func TestGetNamespaceColor(t *testing.T) {
	tests := []struct {
		namespace string
	}{
		{"default"},
		{"kube-system"},
		{"production"},
		{"staging"},
		{"development"},
		{""},
		{"very-long-namespace-name-that-should-still-work"},
	}

	colors := []lipgloss.Color{
		PrimaryBlue, PrimaryGreen, PrimaryYellow, PrimaryRed, PrimaryPurple,
	}

	for _, test := range tests {
		t.Run(test.namespace, func(t *testing.T) {
			result := GetNamespaceColor(test.namespace)

			// Check that result is one of the expected colors
			found := false
			for _, color := range colors {
				if result == color {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("GetNamespaceColor(%q) = %v, want one of %v", test.namespace, result, colors)
			}
		})
	}
}

func TestGetNamespaceColorConsistency(t *testing.T) {
	// Test that the same namespace always returns the same color
	namespace := "test-namespace"

	color1 := GetNamespaceColor(namespace)
	color2 := GetNamespaceColor(namespace)

	if color1 != color2 {
		t.Errorf("GetNamespaceColor should be consistent: got %v and %v for same namespace", color1, color2)
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are not empty
	colorTests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"PrimaryBlue", PrimaryBlue},
		{"PrimaryGreen", PrimaryGreen},
		{"PrimaryYellow", PrimaryYellow},
		{"PrimaryRed", PrimaryRed},
		{"PrimaryPurple", PrimaryPurple},
		{"White", White},
		{"LightGray", LightGray},
		{"Gray", Gray},
		{"DarkGray", DarkGray},
		{"Black", Black},
		{"StatusRunning", StatusRunning},
		{"StatusPending", StatusPending},
		{"StatusFailed", StatusFailed},
		{"StatusSucceeded", StatusSucceeded},
		{"StatusUnknown", StatusUnknown},
		{"StatusTerminated", StatusTerminated},
		{"BorderActive", BorderActive},
		{"BorderInactive", BorderInactive},
		{"Background", Background},
		{"Foreground", Foreground},
		{"Highlight", Highlight},
		{"Selection", Selection},
		{"Error", Error},
		{"Warning", Warning},
		{"Success", Success},
		{"Info", Info},
	}

	for _, test := range colorTests {
		t.Run(test.name, func(t *testing.T) {
			if test.color == "" {
				t.Errorf("Color constant %s should not be empty", test.name)
			}
		})
	}
}

func TestThemeStructure(t *testing.T) {
	// Test that Theme struct can be created and used
	customTheme := Theme{
		Primary:    PrimaryBlue,
		Secondary:  PrimaryGreen,
		Background: Black,
		Foreground: White,
		Border:     Gray,
		Highlight:  PrimaryYellow,
		Error:      PrimaryRed,
		Warning:    PrimaryYellow,
		Success:    PrimaryGreen,
		Info:       PrimaryBlue,
	}

	// Test that all fields are accessible
	if customTheme.Primary == "" {
		t.Error("Custom theme Primary should be accessible")
	}
	if customTheme.Secondary == "" {
		t.Error("Custom theme Secondary should be accessible")
	}
}

func BenchmarkGetStatusColor(b *testing.B) {
	statuses := []string{"Running", "Pending", "Failed", "Succeeded", "Unknown"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetStatusColor(statuses[i%len(statuses)])
	}
}

func BenchmarkGetNamespaceColor(b *testing.B) {
	namespaces := []string{"default", "kube-system", "production", "staging", "development"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetNamespaceColor(namespaces[i%len(namespaces)])
	}
}
