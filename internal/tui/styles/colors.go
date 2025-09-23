package styles

import "github.com/charmbracelet/lipgloss"

// Color constants for the kubeoptic TUI theme
const (
	// Primary colors
	PrimaryBlue   = lipgloss.Color("#61DAFB")
	PrimaryGreen  = lipgloss.Color("#98FB98")
	PrimaryYellow = lipgloss.Color("#F0E68C")
	PrimaryRed    = lipgloss.Color("#FF6B6B")
	PrimaryPurple = lipgloss.Color("#DDA0DD")

	// Grayscale colors
	White     = lipgloss.Color("#FFFFFF")
	LightGray = lipgloss.Color("#E5E5E5")
	Gray      = lipgloss.Color("#808080")
	DarkGray  = lipgloss.Color("#404040")
	Black     = lipgloss.Color("#000000")

	// Status colors
	StatusRunning    = lipgloss.Color("#00FF00")
	StatusPending    = lipgloss.Color("#FFFF00")
	StatusFailed     = lipgloss.Color("#FF0000")
	StatusSucceeded  = lipgloss.Color("#00FF00")
	StatusUnknown    = lipgloss.Color("#808080")
	StatusTerminated = lipgloss.Color("#FFA500")

	// UI element colors
	BorderActive   = lipgloss.Color("#61DAFB")
	BorderInactive = lipgloss.Color("#404040")
	Background     = lipgloss.Color("#1A1A1A")
	Foreground     = lipgloss.Color("#FFFFFF")
	Highlight      = lipgloss.Color("#61DAFB")
	Selection      = lipgloss.Color("#2D2D2D")
	Error          = lipgloss.Color("#FF6B6B")
	Warning        = lipgloss.Color("#F0E68C")
	Success        = lipgloss.Color("#98FB98")
	Info           = lipgloss.Color("#61DAFB")
)

// Theme represents a color theme for the TUI
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Border     lipgloss.Color
	Highlight  lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Success    lipgloss.Color
	Info       lipgloss.Color
}

// DefaultTheme returns the default color theme
func DefaultTheme() Theme {
	return Theme{
		Primary:    PrimaryBlue,
		Secondary:  PrimaryGreen,
		Background: Background,
		Foreground: Foreground,
		Border:     BorderActive,
		Highlight:  Highlight,
		Error:      Error,
		Warning:    Warning,
		Success:    Success,
		Info:       Info,
	}
}

// DarkTheme returns a dark color theme
func DarkTheme() Theme {
	return Theme{
		Primary:    PrimaryPurple,
		Secondary:  PrimaryYellow,
		Background: Black,
		Foreground: White,
		Border:     DarkGray,
		Highlight:  PrimaryPurple,
		Error:      PrimaryRed,
		Warning:    PrimaryYellow,
		Success:    PrimaryGreen,
		Info:       PrimaryBlue,
	}
}

// GetStatusColor returns the appropriate color for a pod status
func GetStatusColor(status string) lipgloss.Color {
	switch status {
	case "Running":
		return StatusRunning
	case "Pending":
		return StatusPending
	case "Failed":
		return StatusFailed
	case "Succeeded":
		return StatusSucceeded
	case "Terminated":
		return StatusTerminated
	default:
		return StatusUnknown
	}
}

// GetNamespaceColor returns a color for namespace based on hash
func GetNamespaceColor(namespace string) lipgloss.Color {
	colors := []lipgloss.Color{
		PrimaryBlue, PrimaryGreen, PrimaryYellow, PrimaryRed, PrimaryPurple,
	}
	hash := 0
	for _, c := range namespace {
		hash += int(c)
	}
	return colors[hash%len(colors)]
}
