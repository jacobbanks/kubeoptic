package styles

import "github.com/charmbracelet/lipgloss"

// Component-specific styles for kubeoptic TUI components

// HeaderStyles returns styles for the application header
func HeaderStyles(theme Theme, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Background).
		Width(width).
		Padding(0, PaddingMedium).
		Border(GetBorders().Round, false, false, true, false).
		BorderForeground(theme.Border)
}

// StatusBarStyles returns styles for the status bar
func StatusBarStyles(theme Theme, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Secondary).
		Foreground(theme.Background).
		Width(width).
		Padding(0, PaddingMedium).
		Bold(true)
}

// ListStyles returns styles for list components
type ListStyles struct {
	Container    lipgloss.Style
	Title        lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	FilteredItem lipgloss.Style
	EmptyState   lipgloss.Style
}

// NewListStyles creates styles for list components
func NewListStyles(theme Theme, width, height int, focused bool) ListStyles {
	borderColor := theme.Border
	if focused {
		borderColor = theme.Primary
	}

	return ListStyles{
		Container: lipgloss.NewStyle().
			Border(GetBorders().Round).
			BorderForeground(borderColor).
			Width(width).
			Height(height).
			Padding(PaddingSmall),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Padding(0, PaddingSmall),

		Item: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			Padding(0, PaddingSmall),

		SelectedItem: lipgloss.NewStyle().
			Background(theme.Highlight).
			Foreground(theme.Background).
			Bold(true).
			Padding(0, PaddingSmall),

		FilteredItem: lipgloss.NewStyle().
			Foreground(theme.Secondary).
			Padding(0, PaddingSmall),

		EmptyState: lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true).
			Align(lipgloss.Center).
			Padding(PaddingLarge),
	}
}

// LogViewerStyles returns styles for the log viewer component
type LogViewerStyles struct {
	Container  lipgloss.Style
	Title      lipgloss.Style
	LogLine    lipgloss.Style
	ErrorLog   lipgloss.Style
	WarningLog lipgloss.Style
	InfoLog    lipgloss.Style
	DebugLog   lipgloss.Style
	Timestamp  lipgloss.Style
	ScrollBar  lipgloss.Style
}

// NewLogViewerStyles creates styles for the log viewer
func NewLogViewerStyles(theme Theme, width, height int, focused bool) LogViewerStyles {
	borderColor := theme.Border
	if focused {
		borderColor = theme.Primary
	}

	return LogViewerStyles{
		Container: lipgloss.NewStyle().
			Border(GetBorders().Round).
			BorderForeground(borderColor).
			Width(width).
			Height(height).
			Padding(PaddingSmall),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Padding(0, PaddingSmall),

		LogLine: lipgloss.NewStyle().
			Foreground(theme.Foreground),

		ErrorLog: lipgloss.NewStyle().
			Foreground(theme.Error),

		WarningLog: lipgloss.NewStyle().
			Foreground(theme.Warning),

		InfoLog: lipgloss.NewStyle().
			Foreground(theme.Info),

		DebugLog: lipgloss.NewStyle().
			Foreground(Gray),

		Timestamp: lipgloss.NewStyle().
			Foreground(Gray).
			Width(20),

		ScrollBar: lipgloss.NewStyle().
			Background(theme.Border).
			Foreground(theme.Primary),
	}
}

// PodStatusStyles returns styles for different pod statuses
func PodStatusStyles(status string, theme Theme) lipgloss.Style {
	color := GetStatusColor(status)
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Padding(0, PaddingSmall)
}

// ContextStyles returns styles for context indicators
func ContextStyles(theme Theme, isActive bool) lipgloss.Style {
	style := lipgloss.NewStyle().Padding(0, PaddingSmall)

	if isActive {
		return style.
			Foreground(theme.Primary).
			Bold(true).
			Background(Selection)
	}

	return style.Foreground(theme.Foreground)
}

// NamespaceStyles returns styles for namespace indicators
func NamespaceStyles(namespace string, theme Theme, isSelected bool) lipgloss.Style {
	color := GetNamespaceColor(namespace)
	style := lipgloss.NewStyle().
		Foreground(color).
		Padding(0, PaddingSmall)

	if isSelected {
		return style.
			Bold(true).
			Background(Selection)
	}

	return style
}

// SearchInputStyles returns styles for search input components
func SearchInputStyles(theme Theme, focused bool) lipgloss.Style {
	borderColor := theme.Border
	if focused {
		borderColor = theme.Primary
	}

	return lipgloss.NewStyle().
		Border(GetBorders().Round).
		BorderForeground(borderColor).
		Foreground(theme.Foreground).
		Background(theme.Background).
		Padding(0, PaddingSmall)
}

// HelpStyles returns styles for help text
func HelpStyles(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Gray).
		Italic(true).
		Padding(PaddingSmall)
}

// ErrorMessageStyles returns styles for error messages
func ErrorMessageStyles(theme Theme, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Error).
		Background(theme.Background).
		Bold(true).
		Border(GetBorders().Round).
		BorderForeground(theme.Error).
		Width(width).
		Padding(PaddingMedium).
		Align(lipgloss.Center)
}
