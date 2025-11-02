package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui/styles"
)

// ViewMode represents different layout modes for the application
type ViewMode int

const (
	ThreePanelView ViewMode = iota // Context | Namespace/Pod | Status
	LogFullScreen                  // Full-screen log view
)

// FocusedPanel represents which panel currently has focus
type FocusedPanel int

const (
	ContextPanel FocusedPanel = iota
	NamespacePanel
	PodPanel
	LogPanel
)

// App represents the main TUI application
type App struct {
	// Core state
	kubeoptic *models.Kubeoptic

	// UI state
	viewMode     ViewMode
	focusedPanel FocusedPanel
	width        int
	height       int

	// Enhanced event handling
	helpVisible bool

	// Components as interfaces to avoid import cycle
	contextList   ComponentRenderer
	namespaceList ComponentRenderer
	podList       ComponentRenderer
	logView       ComponentRenderer
	statusBar     ComponentRenderer

	// Layout
	theme       styles.Theme
	ready       bool
	initialized bool
	err         error
}

// NewApp creates a new TUI application
func NewApp(kubeoptic *models.Kubeoptic) *App {
	theme := styles.DefaultTheme()

	app := &App{
		kubeoptic:    kubeoptic,
		viewMode:     ThreePanelView,
		focusedPanel: ContextPanel,
		theme:        theme,
		ready:        false,
		initialized:  false,
	}

	return app
}

// SetComponents allows external setting of components to avoid import cycles
func (a *App) SetComponents(contextList, namespaceList, podList, logView, statusBar ComponentRenderer) {
	a.contextList = contextList
	a.namespaceList = namespaceList
	a.podList = podList
	a.logView = logView
	a.statusBar = statusBar
}

// Init implements tea.Model interface
func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize each component if they exist
	if a.contextList != nil {
		cmds = append(cmds, a.contextList.Init())
	}
	if a.namespaceList != nil {
		cmds = append(cmds, a.namespaceList.Init())
	}
	if a.podList != nil {
		cmds = append(cmds, a.podList.Init())
	}
	if a.logView != nil {
		cmds = append(cmds, a.logView.Init())
	}
	if a.statusBar != nil {
		cmds = append(cmds, a.statusBar.Init())
	}

	// Mark as initialized
	cmds = append(cmds, func() tea.Msg {
		return InitCompleteMsg{}
	})

	return tea.Batch(cmds...)
}

// Update implements tea.Model interface
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.updateComponentSizes()
		return a, nil

	case tea.KeyMsg:
		// Handle error overlay - any key clears the error
		if a.err != nil {
			a.err = nil
			return a, nil
		}

		// Handle help overlay
		if a.helpVisible {
			switch msg.String() {
			case "?", "esc":
				a.helpVisible = false
				return a, nil
			}
			return a, nil
		}

		// Handle global key bindings
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit

		case "?":
			a.helpVisible = true
			return a, nil

		case "tab":
			a.nextPanel()
			return a, a.updateFocus()

		case "shift+tab":
			a.prevPanel()
			return a, a.updateFocus()

		case "f", "F11":
			// Toggle full-screen log view
			if a.viewMode == ThreePanelView {
				a.viewMode = LogFullScreen
				a.focusedPanel = LogPanel
			} else {
				a.viewMode = ThreePanelView
				a.focusedPanel = PodPanel
			}
			a.updateComponentSizes()
			return a, a.updateFocus()

		case "esc":
			// Exit full-screen mode or go back
			if a.viewMode == LogFullScreen {
				a.viewMode = ThreePanelView
				a.focusedPanel = PodPanel
				a.updateComponentSizes()
				return a, a.updateFocus()
			}
			// Otherwise, navigate back in the workflow
			return a, a.navigateBack()
		}

		// Route key events to focused component
		return a.routeKeyEvent(msg)

	case InitCompleteMsg:
		a.initialized = true
		return a, a.updateFocus()

	case ContextSelectedMsg:
		// Handle context selection
		if msg.Context != nil {
			a.focusedPanel = NamespacePanel
			if a.namespaceList != nil {
				if refreshable, ok := a.namespaceList.(DataProvider); ok {
					cmds = append(cmds, refreshable.RefreshData())
				}
			}
		}
		return a, tea.Batch(cmds...)

	case PodSelectedMsg:
		// Handle pod selection - switch to log view
		if msg.Pod != nil {
			a.viewMode = LogFullScreen
			a.focusedPanel = LogPanel
			a.updateComponentSizes()
			cmds = append(cmds, a.updateFocus())
			// TODO: Start log streaming when LogView component is ready
		}
		return a, tea.Batch(cmds...)

	case ErrorMsg:
		// Handle errors globally
		a.err = msg.Error
		return a, nil

	case ClearStatusMsg:
		a.err = nil
		return a, nil
	}

	// Update components based on focus
	return a.updateComponents(msg)
}

// View implements tea.Model interface
func (a *App) View() string {
	if !a.ready {
		return "Initializing kubeoptic TUI..."
	}

	if a.err != nil {
		return a.renderError()
	}

	// Render help overlay if visible
	if a.helpVisible {
		return a.renderHelpOverlay()
	}

	switch a.viewMode {
	case ThreePanelView:
		return a.renderThreePanelView()
	case LogFullScreen:
		return a.renderLogFullScreen()
	default:
		return "Unknown view mode"
	}
}

// updateComponentSizes adjusts component sizes based on current layout
func (a *App) updateComponentSizes() {
	if !a.ready {
		return
	}

	switch a.viewMode {
	case ThreePanelView:
		// Three-panel layout: Context | NamespacePod | Status
		panelWidth := a.width / 3
		panelHeight := a.height - 3 // Leave space for status bar

		// Update component sizes if they support it
		components := []ComponentRenderer{a.contextList, a.namespaceList, a.podList}
		for _, comp := range components {
			if comp != nil {
				if resizable, ok := comp.(Resizable); ok {
					resizable.SetSize(panelWidth, panelHeight)
				}
			}
		}

	case LogFullScreen:
		// Full-screen log view
		if a.logView != nil {
			if resizable, ok := a.logView.(Resizable); ok {
				resizable.SetSize(a.width, a.height)
			}
		}
	}
}

// updateFocus manages focus between components
func (a *App) updateFocus() tea.Cmd {
	var cmds []tea.Cmd

	// Blur all components first
	components := []ComponentRenderer{a.contextList, a.namespaceList, a.podList, a.logView}
	for _, comp := range components {
		if comp != nil {
			if focusable, ok := comp.(Focusable); ok {
				cmds = append(cmds, focusable.Blur())
			}
		}
	}

	// Focus the active component
	var activeComponent ComponentRenderer
	switch a.focusedPanel {
	case ContextPanel:
		activeComponent = a.contextList
	case NamespacePanel:
		activeComponent = a.namespaceList
	case PodPanel:
		activeComponent = a.podList
	case LogPanel:
		activeComponent = a.logView
	}

	if activeComponent != nil {
		if focusable, ok := activeComponent.(Focusable); ok {
			cmds = append(cmds, focusable.Focus())
		}
	}

	return tea.Batch(cmds...)
}

// nextPanel switches focus to the next panel
func (a *App) nextPanel() {
	switch a.viewMode {
	case ThreePanelView:
		switch a.focusedPanel {
		case ContextPanel:
			a.focusedPanel = NamespacePanel
		case NamespacePanel:
			// Check if we have pods to show
			if len(a.kubeoptic.GetPods()) > 0 {
				a.focusedPanel = PodPanel
			} else {
				a.focusedPanel = ContextPanel
			}
		case PodPanel:
			a.focusedPanel = ContextPanel
		}
	case LogFullScreen:
		// In full-screen mode, stay on log panel
		a.focusedPanel = LogPanel
	}
}

// prevPanel switches focus to the previous panel
func (a *App) prevPanel() {
	switch a.viewMode {
	case ThreePanelView:
		switch a.focusedPanel {
		case ContextPanel:
			if len(a.kubeoptic.GetPods()) > 0 {
				a.focusedPanel = PodPanel
			} else {
				a.focusedPanel = NamespacePanel
			}
		case NamespacePanel:
			a.focusedPanel = ContextPanel
		case PodPanel:
			a.focusedPanel = NamespacePanel
		}
	case LogFullScreen:
		// In full-screen mode, stay on log panel
		a.focusedPanel = LogPanel
	}
}

// navigateBack handles backward navigation in the application flow
func (a *App) navigateBack() tea.Cmd {
	switch a.kubeoptic.GetFocusedView() {
	case models.LogView:
		// Go back from log view to pod view
		a.viewMode = ThreePanelView
		a.focusedPanel = PodPanel
		a.updateComponentSizes()
		return a.updateFocus()

	case models.PodView:
		// Go back from pod view to namespace view
		a.focusedPanel = NamespacePanel
		return a.updateFocus()

	case models.NamespaceView:
		// Go back from namespace view to context view
		a.focusedPanel = ContextPanel
		return a.updateFocus()

	default:
		// Already at top level
		return nil
	}
}

// routeKeyEvent routes key events to the currently focused component
func (a *App) routeKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	var activeComponent ComponentRenderer
	switch a.focusedPanel {
	case ContextPanel:
		activeComponent = a.contextList
	case NamespacePanel:
		activeComponent = a.namespaceList
	case PodPanel:
		activeComponent = a.podList
	case LogPanel:
		activeComponent = a.logView
	}

	if activeComponent != nil {
		if eventHandler, ok := activeComponent.(EventHandler); ok {
			_, cmd = eventHandler.HandleKeyEvent(msg)
		} else {
			// Fallback to direct Update if EventHandler not implemented
			_, cmd = activeComponent.Update(msg)
		}
	}

	return a, cmd
}

// updateComponents updates all components with the given message
func (a *App) updateComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	components := []ComponentRenderer{a.contextList, a.namespaceList, a.podList, a.logView, a.statusBar}
	for _, comp := range components {
		if comp != nil {
			_, cmd := comp.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return a, tea.Batch(cmds...)
}

// renderThreePanelView renders the three-panel layout
func (a *App) renderThreePanelView() string {
	if !a.initialized {
		return "Loading components..."
	}

	// Get panel views
	contextView := ""
	if a.contextList != nil {
		contextView = a.contextList.View()
	}

	middleView := ""
	// Show namespace list by default, or pod list if namespace is selected
	if a.kubeoptic.GetSelectedNamespace() != "" && len(a.kubeoptic.GetPods()) > 0 {
		if a.podList != nil {
			middleView = a.podList.View()
		}
	} else {
		if a.namespaceList != nil {
			middleView = a.namespaceList.View()
		}
	}

	rightView := ""
	if a.statusBar != nil {
		rightView = a.statusBar.View()
	}

	statusView := a.renderStatusBar()

	// Create layout using lipgloss
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		a.addPanelBorder(contextView, a.focusedPanel == ContextPanel),
		a.addPanelBorder(middleView, a.focusedPanel == NamespacePanel || a.focusedPanel == PodPanel),
		a.addPanelBorder(rightView, false), // Third panel
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		statusView,
	)
}

// renderLogFullScreen renders the full-screen log view
func (a *App) renderLogFullScreen() string {
	if a.logView != nil {
		return a.logView.View()
	}

	// Fallback content when LogView component is not ready
	logContent := "Log View (Full Screen)\n[Log content would appear here]\n\nPress 'f' or Esc to return to main view"

	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Primary).
		Render(logContent)
}

// renderStatusBar renders the status bar
func (a *App) renderStatusBar() string {
	// Basic status bar implementation
	leftStatus := fmt.Sprintf("Context: %s | Namespace: %s",
		a.kubeoptic.GetSelectedContext(),
		a.kubeoptic.GetSelectedNamespace())

	rightStatus := "Tab: switch panels | f: full-screen | q: quit"

	statusStyle := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Background(styles.DarkGray).
		Width(a.width).
		Padding(0, 1)

	// Create status bar with left and right aligned content
	gap := a.width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus) - 4
	if gap < 0 {
		gap = 0
	}

	statusContent := leftStatus + lipgloss.NewStyle().Width(gap).Render("") + rightStatus

	return statusStyle.Render(statusContent)
}

// renderError renders error messages
func (a *App) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(a.theme.Error).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Error).
		Padding(1).
		Width(a.width - 4)

	content := fmt.Sprintf("Error: %s\n\nPress any key to continue", a.err.Error())

	return errorStyle.Render(content)
}

// addPanelBorder adds a border to a panel with focus indication
func (a *App) addPanelBorder(content string, focused bool) string {
	borderColor := styles.BorderInactive
	if focused {
		borderColor = a.theme.Primary
	}

	panelWidth := a.width/3 - 2
	panelHeight := a.height - 3

	// Ensure minimum dimensions
	if panelWidth < 10 {
		panelWidth = 10
	}
	if panelHeight < 5 {
		panelHeight = 5
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(panelWidth).
		Height(panelHeight).
		Render(content)
}

// AppStateManager interface implementation
func (a *App) GetCurrentView() models.ViewType {
	return a.kubeoptic.GetFocusedView()
}

func (a *App) SetCurrentView(view models.ViewType) {
	// This would require extending the Kubeoptic model
	// For now, we manage view state in the App
}

func (a *App) GetSelectedContext() string {
	return a.kubeoptic.GetSelectedContext()
}

func (a *App) SetSelectedContext(context string) {
	a.kubeoptic.SelectContext(context)
}

func (a *App) GetSelectedNamespace() string {
	return a.kubeoptic.GetSelectedNamespace()
}

func (a *App) SetSelectedNamespace(namespace string) {
	a.kubeoptic.SelectNamespace(namespace)
}

func (a *App) GetSelectedPod() *services.Pod {
	return a.kubeoptic.GetSelectedPod()
}

func (a *App) SetSelectedPod(pod *services.Pod) {
	if pod != nil {
		a.kubeoptic.SelectPod(pod.Name)
	}
}

func (a *App) GetPodSearchQuery() string {
	return a.kubeoptic.GetSearchQuery()
}

func (a *App) SetPodSearchQuery(query string) {
	a.kubeoptic.SearchPods(query)
}

func (a *App) GetFilteredPods() []services.Pod {
	pods := a.kubeoptic.GetPods()
	if pods == nil {
		return []services.Pod{}
	}
	return pods
}

func (a *App) GetLogBuffer() []string {
	logs := a.kubeoptic.GetLogBuffer()
	if logs == nil {
		return []string{}
	}
	return logs
}

func (a *App) IsFollowing() bool {
	return a.kubeoptic.IsFollowing()
}

func (a *App) SetFollowing(following bool) {
	// TODO: Implement in Kubeoptic model
}

// renderHelpOverlay renders the help overlay
func (a *App) renderHelpOverlay() string {
	content := []string{
		lipgloss.NewStyle().Bold(true).Foreground(a.theme.Primary).Render("Help - kubeoptic"),
		"",
		lipgloss.NewStyle().Foreground(a.theme.Secondary).Italic(true).Render("Current View: " + a.getCurrentViewName()),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(a.theme.Secondary).Render("Global"),
		a.formatKeyBinding("?", "help"),
		a.formatKeyBinding("q, ctrl+c", "quit"),
		a.formatKeyBinding("tab", "next panel"),
		a.formatKeyBinding("shift+tab", "prev panel"),
		a.formatKeyBinding("f", "toggle fullscreen"),
		a.formatKeyBinding("esc", "back"),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(a.theme.Secondary).Render("Navigation"),
		a.formatKeyBinding("↑/↓, j/k", "navigate"),
		a.formatKeyBinding("enter", "select"),
		a.formatKeyBinding("/", "search"),
		"",
		"Press '?' or 'esc' to close help",
	}

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Primary).
		Background(a.theme.Background).
		Padding(1, 2).
		Width(a.width - 8).
		Height(a.height - 4)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center,
		helpStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...)))
}

// getCurrentViewName returns the current view name for help display
func (a *App) getCurrentViewName() string {
	switch a.kubeoptic.GetFocusedView() {
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

// formatKeyBinding formats a key binding for display
func (a *App) formatKeyBinding(keys, desc string) string {
	keyStyle := lipgloss.NewStyle().Foreground(a.theme.Primary).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(a.theme.Foreground)

	return fmt.Sprintf("  %s  %s",
		keyStyle.Width(12).Render(keys),
		descStyle.Render(desc))
}
