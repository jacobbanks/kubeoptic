package styles

import "github.com/charmbracelet/lipgloss"

// Layout constants for consistent spacing and sizing
const (
	// Padding and margins
	PaddingSmall  = 1
	PaddingMedium = 2
	PaddingLarge  = 4

	MarginSmall  = 1
	MarginMedium = 2
	MarginLarge  = 4

	// Border constants
	BorderWidth   = 1
	BorderRounded = true
	BorderNormal  = false
	BorderThick   = 2

	// Component dimensions
	MinListHeight      = 10
	MinListWidth       = 30
	StatusBarHeight    = 1
	HeaderHeight       = 3
	MinLogViewerHeight = 15
	MinLogViewerWidth  = 50

	// Responsive breakpoints
	SmallScreenWidth  = 80
	MediumScreenWidth = 120
	LargeScreenWidth  = 160
)

// LayoutConfig holds responsive layout configuration
type LayoutConfig struct {
	Width              int
	Height             int
	IsSmallScreen      bool
	IsMediumScreen     bool
	IsLargeScreen      bool
	ContextListWidth   int
	NamespaceListWidth int
	PodListWidth       int
	LogViewerWidth     int
	LogViewerHeight    int
}

// NewLayoutConfig creates a layout configuration for given screen dimensions
func NewLayoutConfig(width, height int) LayoutConfig {
	config := LayoutConfig{
		Width:  width,
		Height: height,
	}

	// Determine screen size category
	config.IsSmallScreen = width <= SmallScreenWidth
	config.IsMediumScreen = width > SmallScreenWidth && width <= MediumScreenWidth
	config.IsLargeScreen = width > MediumScreenWidth

	// Calculate responsive widths
	if config.IsSmallScreen {
		// Small screens: stack vertically or use minimal widths
		config.ContextListWidth = width - 4
		config.NamespaceListWidth = width - 4
		config.PodListWidth = width - 4
		config.LogViewerWidth = width - 4
	} else if config.IsMediumScreen {
		// Medium screens: 3-column layout
		config.ContextListWidth = width / 4
		config.NamespaceListWidth = width / 4
		config.PodListWidth = width / 2
		config.LogViewerWidth = width - 4
	} else {
		// Large screens: optimized 3-column layout
		config.ContextListWidth = width / 5
		config.NamespaceListWidth = width / 5
		config.PodListWidth = (width * 3) / 5
		config.LogViewerWidth = width - 4
	}

	// Calculate log viewer height (reserve space for header and status)
	config.LogViewerHeight = height - HeaderHeight - StatusBarHeight - 2

	return config
}

// GetPanelWidth returns the appropriate panel width for the current layout
func (lc LayoutConfig) GetPanelWidth() int {
	if lc.IsSmallScreen {
		return lc.Width - 4
	}
	return lc.Width / 3
}

// GetMainViewHeight returns the height available for the main content area
func (lc LayoutConfig) GetMainViewHeight() int {
	return lc.Height - HeaderHeight - StatusBarHeight
}

// GetListHeight returns the appropriate height for list components
func (lc LayoutConfig) GetListHeight() int {
	mainHeight := lc.GetMainViewHeight()
	if mainHeight < MinListHeight {
		return MinListHeight
	}
	return mainHeight - 2 // Reserve space for borders
}

// ShouldStackVertically returns true if components should stack vertically
func (lc LayoutConfig) ShouldStackVertically() bool {
	return lc.IsSmallScreen
}

// CalculateThreePanelLayout returns widths for a three-panel horizontal layout
func (lc LayoutConfig) CalculateThreePanelLayout() (int, int, int) {
	if lc.IsSmallScreen {
		// On small screens, return full width for stacking
		fullWidth := lc.Width - 2
		return fullWidth, fullWidth, fullWidth
	}

	// Calculate proportional widths for three panels
	availableWidth := lc.Width - 6 // Account for borders and spacing
	leftWidth := availableWidth / 4
	middleWidth := availableWidth / 3
	rightWidth := availableWidth - leftWidth - middleWidth

	return leftWidth, middleWidth, rightWidth
}

// GetBorderStyle returns the appropriate border style for the layout
func (lc LayoutConfig) GetBorderStyle() lipgloss.Border {
	if lc.IsSmallScreen {
		return lipgloss.NormalBorder()
	}
	return lipgloss.RoundedBorder()
}

// GetBorders returns common border styles
func GetBorders() struct {
	None   lipgloss.Border
	Round  lipgloss.Border
	Double lipgloss.Border
	Dotted lipgloss.Border
} {
	return struct {
		None   lipgloss.Border
		Round  lipgloss.Border
		Double lipgloss.Border
		Dotted lipgloss.Border
	}{
		None:   lipgloss.NormalBorder(),
		Round:  lipgloss.RoundedBorder(),
		Double: lipgloss.DoubleBorder(),
		Dotted: lipgloss.Border{},
	}
}
