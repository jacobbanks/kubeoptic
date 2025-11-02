package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"kubeoptic/internal/models"
	"kubeoptic/internal/tui/styles"
)

// StatusBar represents the status bar component
type StatusBar struct {
	width            int
	height           int
	theme            styles.Theme
	kubeoptic        *models.Kubeoptic
	connectionStatus string
}

// NewStatusBar creates a new status bar component
func NewStatusBar(theme styles.Theme, kubeoptic *models.Kubeoptic) *StatusBar {
	return &StatusBar{
		width:            0,
		height:           1,
		theme:            theme,
		kubeoptic:        kubeoptic,
		connectionStatus: "Connected",
	}
}

// SetSize updates the status bar dimensions
func (s *StatusBar) SetSize(width, height int) {
	s.width = width
	if height > 0 {
		s.height = height
	}
}

// SetConnectionStatus updates the connection status
func (s *StatusBar) SetConnectionStatus(status string) {
	s.connectionStatus = status
}

// Init implements tea.Model interface
func (s *StatusBar) Init() tea.Cmd {
	return nil
}

// Update handles tea.Msg updates
func (s *StatusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)
	}
	return tea.Model(s), nil
}

// View renders the status bar
func (s *StatusBar) View() string {
	if s.width <= 0 {
		return ""
	}

	// Build status sections
	leftSection := s.buildLeftSection()
	rightSection := s.buildRightSection()

	// Calculate available space for center section
	usedSpace := lipgloss.Width(leftSection) + lipgloss.Width(rightSection)
	availableSpace := s.width - usedSpace

	// Build center section with available space
	centerSection := s.buildCenterSection(availableSpace)

	// Combine sections
	statusLine := leftSection + centerSection + rightSection

	// Apply status bar styling
	style := styles.StatusBarStyles(s.theme, s.width)
	return style.Render(statusLine)
}

// buildLeftSection creates the left side of the status bar (context info)
func (s *StatusBar) buildLeftSection() string {
	if s.kubeoptic == nil {
		return ""
	}

	var parts []string

	// Context information
	if context := s.kubeoptic.GetSelectedContext(); context != "" {
		contextStyle := lipgloss.NewStyle().
			Foreground(s.theme.Primary).
			Bold(true)
		parts = append(parts, fmt.Sprintf("ctx:%s", contextStyle.Render(context)))
	}

	// Namespace information
	if namespace := s.kubeoptic.GetSelectedNamespace(); namespace != "" {
		namespaceStyle := lipgloss.NewStyle().
			Foreground(styles.GetNamespaceColor(namespace)).
			Bold(true)
		parts = append(parts, fmt.Sprintf("ns:%s", namespaceStyle.Render(namespace)))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " │ ") + " │ "
}

// buildCenterSection creates the center of the status bar (pod info and counts)
func (s *StatusBar) buildCenterSection(availableWidth int) string {
	if s.kubeoptic == nil || availableWidth <= 0 {
		return strings.Repeat(" ", max(0, availableWidth))
	}

	var centerParts []string

	// Selected pod information
	if pod := s.kubeoptic.GetSelectedPod(); pod != nil {
		podStyle := lipgloss.NewStyle().
			Foreground(styles.GetStatusColor(string(pod.Status))).
			Bold(true)
		centerParts = append(centerParts, fmt.Sprintf("pod:%s", podStyle.Render(pod.Name)))
	}

	// Pod count information
	if count := s.kubeoptic.GetPodCount(); count != "" {
		countStyle := lipgloss.NewStyle().
			Foreground(s.theme.Info)
		centerParts = append(centerParts, countStyle.Render(count))
	}

	// Search query if active
	if query := s.kubeoptic.GetSearchQuery(); query != "" {
		searchStyle := lipgloss.NewStyle().
			Foreground(s.theme.Warning).
			Italic(true)
		centerParts = append(centerParts, fmt.Sprintf("search:%s", searchStyle.Render(query)))
	}

	centerText := strings.Join(centerParts, " │ ")
	centerWidth := lipgloss.Width(centerText)

	// If center text is too wide, truncate intelligently
	if centerWidth > availableWidth {
		return s.truncateCenter(centerText, availableWidth)
	}

	// Center the text within available space
	padding := (availableWidth - centerWidth) / 2
	leftPadding := strings.Repeat(" ", padding)
	rightPadding := strings.Repeat(" ", availableWidth-centerWidth-padding)

	return leftPadding + centerText + rightPadding
}

// buildRightSection creates the right side of the status bar (connection status)
func (s *StatusBar) buildRightSection() string {
	if s.connectionStatus == "" {
		return ""
	}

	var statusStyle lipgloss.Style
	switch s.connectionStatus {
	case "Connected":
		statusStyle = lipgloss.NewStyle().
			Foreground(s.theme.Success).
			Bold(true)
	case "Connecting":
		statusStyle = lipgloss.NewStyle().
			Foreground(s.theme.Warning).
			Bold(true)
	case "Disconnected", "Error":
		statusStyle = lipgloss.NewStyle().
			Foreground(s.theme.Error).
			Bold(true)
	default:
		statusStyle = lipgloss.NewStyle().
			Foreground(s.theme.Foreground)
	}

	return " │ " + statusStyle.Render(s.connectionStatus)
}

// truncateCenter intelligently truncates the center section when space is limited
func (s *StatusBar) truncateCenter(text string, maxWidth int) string {
	if maxWidth <= 3 {
		return strings.Repeat(" ", maxWidth)
	}

	// Try to preserve the most important information (pod name if present)
	if strings.Contains(text, "pod:") {
		parts := strings.Split(text, " │ ")
		for _, part := range parts {
			if strings.HasPrefix(part, "pod:") {
				if lipgloss.Width(part) <= maxWidth-3 {
					return part + "..." + strings.Repeat(" ", maxWidth-lipgloss.Width(part)-3)
				}
			}
		}
	}

	// Fallback: simple truncation
	if len(text) > maxWidth-3 {
		return text[:maxWidth-3] + "..."
	}

	return text + strings.Repeat(" ", maxWidth-lipgloss.Width(text))
}

// Helper function for max (Go 1.21+ has this built-in)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
