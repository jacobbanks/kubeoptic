package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/models"
	"kubeoptic/internal/tui/keys"
	"kubeoptic/internal/tui/navigation"
	"kubeoptic/internal/tui/styles"
)

// HelpModel represents the help system state
type HelpModel struct {
	visible      bool
	currentView  models.ViewType
	focusState   navigation.FocusState
	keyMap       keys.KeyMap
	width        int
	height       int
	showAdvanced bool
}

// NewHelpModel creates a new help model
func NewHelpModel(keyMap keys.KeyMap) *HelpModel {
	return &HelpModel{
		visible:      false,
		keyMap:       keyMap,
		showAdvanced: false,
	}
}

// IsVisible returns whether the help is currently visible
func (h *HelpModel) IsVisible() bool {
	return h.visible
}

// Show displays the help
func (h *HelpModel) Show() {
	h.visible = true
}

// Hide conceals the help
func (h *HelpModel) Hide() {
	h.visible = false
}

// Toggle toggles the help visibility
func (h *HelpModel) Toggle() {
	h.visible = !h.visible
}

// Update updates the help model with current context
func (h *HelpModel) Update(view models.ViewType, focus navigation.FocusState, width, height int) {
	h.currentView = view
	h.focusState = focus
	h.width = width
	h.height = height
}

// ToggleAdvanced toggles between basic and advanced help
func (h *HelpModel) ToggleAdvanced() {
	h.showAdvanced = !h.showAdvanced
}

// HandleKeyPress handles key presses within the help system
func (h *HelpModel) HandleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, h.keyMap.Help):
		h.Toggle()
		return nil
	case key.Matches(msg, h.keyMap.Back):
		h.Hide()
		return nil
	case msg.String() == "a":
		h.ToggleAdvanced()
		return nil
	}
	return nil
}

// View renders the help system
func (h *HelpModel) View() string {
	if !h.visible {
		return ""
	}

	// Get theme
	theme := styles.DefaultTheme()

	// Create help content
	var content strings.Builder

	// Header
	title := "Help - kubeoptic"
	if h.showAdvanced {
		title += " (Advanced)"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Background).
		Padding(0, 1).
		Margin(1, 0)

	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n\n")

	// Current context info
	contextInfo := h.getContextInfo()
	contextStyle := lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Italic(true).
		Padding(0, 1)

	content.WriteString(contextStyle.Render(contextInfo))
	content.WriteString("\n\n")

	// Key bindings
	keyBindings := h.getKeyBindings()
	content.WriteString(keyBindings)

	// Footer
	footer := h.getFooter()
	content.WriteString("\n")
	content.WriteString(footer)

	// Create help box
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 2).
		Width(h.width - 4).
		Height(h.height - 4)

	return helpStyle.Render(content.String())
}

// getContextInfo returns information about the current context
func (h *HelpModel) getContextInfo() string {
	var context strings.Builder

	// Current view
	context.WriteString(fmt.Sprintf("Current View: %s", h.getViewName()))

	// Current focus (if in multi-panel view)
	if h.currentView == models.NamespaceView || h.currentView == models.PodView {
		context.WriteString(fmt.Sprintf(" | Focus: %s", h.getFocusName()))
	}

	return context.String()
}

// getViewName returns a human-readable view name
func (h *HelpModel) getViewName() string {
	switch h.currentView {
	case models.ContextView:
		return "Contexts"
	case models.NamespaceView:
		return "Namespaces"
	case models.PodView:
		return "Pods"
	case models.LogView:
		return "Logs"
	default:
		return "Unknown"
	}
}

// getFocusName returns a human-readable focus name
func (h *HelpModel) getFocusName() string {
	switch h.focusState {
	case navigation.FocusContext:
		return "Context Panel"
	case navigation.FocusNamespace:
		return "Namespace Panel"
	case navigation.FocusPod:
		return "Pod Panel"
	case navigation.FocusSearch:
		return "Search"
	case navigation.FocusLog:
		return "Log Viewer"
	case navigation.FocusStatusBar:
		return "Status Bar"
	default:
		return "Unknown"
	}
}

// getKeyBindings returns formatted key bindings for the current context
func (h *HelpModel) getKeyBindings() string {
	bindings := keys.GetKeyHelp(h.currentView, h.keyMap)

	var content strings.Builder

	// Global bindings (always available)
	content.WriteString(h.formatKeySection("Global", [][]key.Binding{
		{h.keyMap.Help, h.keyMap.Quit},
	}))

	// View-specific bindings
	viewName := h.getViewName()
	content.WriteString(h.formatKeySection(viewName, bindings))

	// Advanced bindings (if enabled)
	if h.showAdvanced {
		content.WriteString(h.getAdvancedBindings())
	}

	return content.String()
}

// formatKeySection formats a section of key bindings
func (h *HelpModel) formatKeySection(title string, bindings [][]key.Binding) string {
	if len(bindings) == 0 {
		return ""
	}

	theme := styles.DefaultTheme()

	var content strings.Builder

	// Section title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Secondary).
		MarginTop(1)

	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")

	// Key bindings
	keyStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)

	for _, row := range bindings {
		for _, binding := range row {
			if binding.Enabled() {
				keys := strings.Join(binding.Keys(), ", ")
				help := binding.Help().Desc

				line := fmt.Sprintf("  %s  %s",
					keyStyle.Width(12).Render(keys),
					descStyle.Render(help))

				content.WriteString(line)
				content.WriteString("\n")
			}
		}
	}

	return content.String()
}

// getAdvancedBindings returns advanced key bindings
func (h *HelpModel) getAdvancedBindings() string {
	var content strings.Builder

	content.WriteString(h.formatKeySection("Advanced", [][]key.Binding{
		{
			key.NewBinding(
				key.WithKeys("ctrl+u"),
				key.WithHelp("ctrl+u", "scroll up (logs)"),
			),
			key.NewBinding(
				key.WithKeys("ctrl+d"),
				key.WithHelp("ctrl+d", "scroll down (logs)"),
			),
		},
		{
			key.NewBinding(
				key.WithKeys("w"),
				key.WithHelp("w", "toggle word wrap (logs)"),
			),
			key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "save logs to file"),
			),
		},
		{
			key.NewBinding(
				key.WithKeys("g"),
				key.WithHelp("g", "go to top"),
			),
			key.NewBinding(
				key.WithKeys("G"),
				key.WithHelp("G", "go to bottom"),
			),
		},
	}))

	return content.String()
}

// getFooter returns the help footer
func (h *HelpModel) getFooter() string {
	theme := styles.DefaultTheme()

	footerStyle := lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Italic(true).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(theme.Border).
		PaddingTop(1).
		MarginTop(1)

	var footer strings.Builder

	if h.showAdvanced {
		footer.WriteString("Press 'a' for basic help")
	} else {
		footer.WriteString("Press 'a' for advanced help")
	}

	footer.WriteString(" | Press '?' or 'esc' to close")

	return footerStyle.Render(footer.String())
}

// GetQuickHelp returns a quick help string for the status bar
func (h *HelpModel) GetQuickHelp() string {
	switch h.currentView {
	case models.ContextView:
		return "↑/↓ navigate • enter select • / search • ? help • q quit"
	case models.NamespaceView:
		return "↑/↓ navigate • tab switch panel • enter select • esc back • ? help"
	case models.PodView:
		return "↑/↓ navigate • tab switch panel • enter logs • / search • esc back"
	case models.LogView:
		return "↑/↓ scroll • f follow • ctrl+u/d page • esc back • ? help"
	default:
		return "? help • q quit"
	}
}

// HelpOverlay creates a help overlay that can be displayed over other content
type HelpOverlay struct {
	help   *HelpModel
	active bool
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(keyMap keys.KeyMap) *HelpOverlay {
	return &HelpOverlay{
		help:   NewHelpModel(keyMap),
		active: false,
	}
}

// Toggle toggles the help overlay
func (ho *HelpOverlay) Toggle() tea.Cmd {
	ho.active = !ho.active
	if ho.active {
		ho.help.Show()
	} else {
		ho.help.Hide()
	}
	return nil
}

// Update updates the help overlay
func (ho *HelpOverlay) Update(msg tea.Msg, view models.ViewType, focus navigation.FocusState, width, height int) tea.Cmd {
	ho.help.Update(view, focus, width, height)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if ho.active {
			return ho.help.HandleKeyPress(msg)
		}
	case keys.ShowHelpMsg:
		ho.active = true
		ho.help.Show()
	case keys.HideHelpMsg:
		ho.active = false
		ho.help.Hide()
	}

	return nil
}

// View renders the help overlay
func (ho *HelpOverlay) View() string {
	if !ho.active {
		return ""
	}
	return ho.help.View()
}

// IsActive returns whether the help overlay is active
func (ho *HelpOverlay) IsActive() bool {
	return ho.active
}

// GetQuickHelp returns quick help for the status bar
func (ho *HelpOverlay) GetQuickHelp() string {
	return ho.help.GetQuickHelp()
}
