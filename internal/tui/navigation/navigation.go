package navigation

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/keys"
)

// FocusState represents which component has focus in multi-panel views
type FocusState int

const (
	FocusContext FocusState = iota
	FocusNamespace
	FocusPod
	FocusSearch
	FocusLog
	FocusStatusBar
)

// NavigationState manages the current navigation state of the application
type NavigationState struct {
	currentView       models.ViewType
	focusedPanel      FocusState
	previousView      models.ViewType
	helpVisible       bool
	searchMode        bool
	navigationHistory []models.ViewType
}

// NewNavigationState creates a new navigation state starting with context view
func NewNavigationState() *NavigationState {
	return &NavigationState{
		currentView:       models.ContextView,
		focusedPanel:      FocusContext,
		previousView:      models.ContextView,
		helpVisible:       false,
		searchMode:        false,
		navigationHistory: []models.ViewType{models.ContextView},
	}
}

// GetCurrentView returns the current view
func (n *NavigationState) GetCurrentView() models.ViewType {
	return n.currentView
}

// GetFocusedPanel returns the currently focused panel
func (n *NavigationState) GetFocusedPanel() FocusState {
	return n.focusedPanel
}

// IsHelpVisible returns whether help is currently visible
func (n *NavigationState) IsHelpVisible() bool {
	return n.helpVisible
}

// IsSearchMode returns whether we're in search mode
func (n *NavigationState) IsSearchMode() bool {
	return n.searchMode
}

// NavigateToView changes the current view and updates navigation history
func (n *NavigationState) NavigateToView(view models.ViewType) tea.Cmd {
	if n.currentView != view {
		n.previousView = n.currentView
		n.currentView = view
		n.navigationHistory = append(n.navigationHistory, view)

		// Set appropriate focus for the new view
		switch view {
		case models.ContextView:
			n.focusedPanel = FocusContext
		case models.NamespaceView:
			n.focusedPanel = FocusNamespace
		case models.PodView:
			n.focusedPanel = FocusPod
		case models.LogView:
			n.focusedPanel = FocusLog
		}

		// Clear search mode when changing views
		n.searchMode = false

		return func() tea.Msg {
			return tui.FocusChangedMsg{
				Component: n.getFocusComponentName(),
				Focused:   true,
			}
		}
	}
	return nil
}

// NavigateBack goes back to the previous view
func (n *NavigationState) NavigateBack() tea.Cmd {
	var targetView models.ViewType

	// Determine where to go back to based on current view
	switch n.currentView {
	case models.LogView:
		targetView = models.PodView
	case models.PodView:
		targetView = models.NamespaceView
	case models.NamespaceView:
		targetView = models.ContextView
	case models.ContextView:
		// Can't go back from context view
		return nil
	}

	return n.NavigateToView(targetView)
}

// NextPanel moves focus to the next panel in multi-panel views
func (n *NavigationState) NextPanel() tea.Cmd {
	switch n.currentView {
	case models.ContextView, models.LogView:
		// Single panel views - no panel switching
		return nil

	case models.NamespaceView:
		// Context and Namespace panels
		switch n.focusedPanel {
		case FocusContext:
			n.focusedPanel = FocusNamespace
		case FocusNamespace:
			n.focusedPanel = FocusContext
		}

	case models.PodView:
		// Context, Namespace, and Pod panels
		switch n.focusedPanel {
		case FocusContext:
			n.focusedPanel = FocusNamespace
		case FocusNamespace:
			n.focusedPanel = FocusPod
		case FocusPod:
			n.focusedPanel = FocusContext
		}
	}

	return func() tea.Msg {
		return tui.FocusChangedMsg{
			Component: n.getFocusComponentName(),
			Focused:   true,
		}
	}
}

// PrevPanel moves focus to the previous panel in multi-panel views
func (n *NavigationState) PrevPanel() tea.Cmd {
	switch n.currentView {
	case models.ContextView, models.LogView:
		// Single panel views - no panel switching
		return nil

	case models.NamespaceView:
		// Context and Namespace panels
		switch n.focusedPanel {
		case FocusContext:
			n.focusedPanel = FocusNamespace
		case FocusNamespace:
			n.focusedPanel = FocusContext
		}

	case models.PodView:
		// Context, Namespace, and Pod panels
		switch n.focusedPanel {
		case FocusContext:
			n.focusedPanel = FocusPod
		case FocusNamespace:
			n.focusedPanel = FocusContext
		case FocusPod:
			n.focusedPanel = FocusNamespace
		}
	}

	return func() tea.Msg {
		return tui.FocusChangedMsg{
			Component: n.getFocusComponentName(),
			Focused:   true,
		}
	}
}

// ToggleHelp toggles the help display
func (n *NavigationState) ToggleHelp() tea.Cmd {
	n.helpVisible = !n.helpVisible

	if n.helpVisible {
		return func() tea.Msg {
			return keys.ShowHelpMsg{}
		}
	}
	return func() tea.Msg {
		return keys.HideHelpMsg{}
	}
}

// EnterSearchMode puts the application in search mode
func (n *NavigationState) EnterSearchMode() tea.Cmd {
	n.searchMode = true
	n.focusedPanel = FocusSearch

	return func() tea.Msg {
		return tui.SearchQueryChangedMsg{Query: ""}
	}
}

// ExitSearchMode exits search mode
func (n *NavigationState) ExitSearchMode() tea.Cmd {
	n.searchMode = false

	// Restore focus to appropriate panel
	switch n.currentView {
	case models.ContextView:
		n.focusedPanel = FocusContext
	case models.NamespaceView:
		n.focusedPanel = FocusNamespace
	case models.PodView:
		n.focusedPanel = FocusPod
	case models.LogView:
		n.focusedPanel = FocusLog
	}

	return func() tea.Msg {
		return tui.ClearSearchMsg{}
	}
}

// CanNavigateForward checks if forward navigation is possible
func (n *NavigationState) CanNavigateForward() bool {
	switch n.currentView {
	case models.ContextView:
		return true // Can go to namespace view
	case models.NamespaceView:
		return true // Can go to pod view
	case models.PodView:
		return true // Can go to log view
	case models.LogView:
		return false // Can't go forward from log view
	}
	return false
}

// CanNavigateBack checks if back navigation is possible
func (n *NavigationState) CanNavigateBack() bool {
	return n.currentView != models.ContextView
}

// GetBreadcrumb returns a breadcrumb string for the current navigation state
func (n *NavigationState) GetBreadcrumb() string {
	switch n.currentView {
	case models.ContextView:
		return "Contexts"
	case models.NamespaceView:
		return "Contexts > Namespaces"
	case models.PodView:
		return "Contexts > Namespaces > Pods"
	case models.LogView:
		return "Contexts > Namespaces > Pods > Logs"
	}
	return ""
}

// HandleNavigation processes navigation-related key presses
func (n *NavigationState) HandleNavigation(msg tea.KeyMsg, keyMap keys.KeyMap) tea.Cmd {
	switch {
	case key.Matches(msg, keyMap.Back):
		return n.NavigateBack()

	case key.Matches(msg, keyMap.NextPanel):
		return n.NextPanel()

	case key.Matches(msg, keyMap.PrevPanel):
		return n.PrevPanel()

	case key.Matches(msg, keyMap.Search):
		return n.EnterSearchMode()

	case key.Matches(msg, keyMap.Help):
		return n.ToggleHelp()

	case key.Matches(msg, keyMap.SwitchContext):
		return n.NavigateToView(models.ContextView)

	case key.Matches(msg, keyMap.SelectNamespace):
		return n.NavigateToView(models.NamespaceView)

	case key.Matches(msg, keyMap.SelectPod):
		return n.NavigateToView(models.PodView)
	}

	return nil
}

// getFocusComponentName returns the name of the currently focused component
func (n *NavigationState) getFocusComponentName() string {
	switch n.focusedPanel {
	case FocusContext:
		return "context"
	case FocusNamespace:
		return "namespace"
	case FocusPod:
		return "pod"
	case FocusSearch:
		return "search"
	case FocusLog:
		return "log"
	case FocusStatusBar:
		return "status"
	}
	return "unknown"
}

// ViewTransition represents a view transition with validation
type ViewTransition struct {
	From         models.ViewType
	To           models.ViewType
	Condition    func() bool
	OnTransition func() tea.Cmd
}

// ValidTransitions defines the allowed view transitions
var ValidTransitions = []ViewTransition{
	{
		From:      models.ContextView,
		To:        models.NamespaceView,
		Condition: func() bool { return true },
	},
	{
		From:      models.NamespaceView,
		To:        models.PodView,
		Condition: func() bool { return true },
	},
	{
		From:      models.PodView,
		To:        models.LogView,
		Condition: func() bool { return true },
	},
	{
		From:      models.LogView,
		To:        models.PodView,
		Condition: func() bool { return true },
	},
	{
		From:      models.PodView,
		To:        models.NamespaceView,
		Condition: func() bool { return true },
	},
	{
		From:      models.NamespaceView,
		To:        models.ContextView,
		Condition: func() bool { return true },
	},
}

// IsValidTransition checks if a view transition is valid
func IsValidTransition(from, to models.ViewType) bool {
	for _, transition := range ValidTransitions {
		if transition.From == from && transition.To == to {
			if transition.Condition != nil {
				return transition.Condition()
			}
			return true
		}
	}
	return false
}

// Navigation messages for tea.Update patterns
type NavigationMsg struct {
	Action string
	Data   interface{}
}

// Navigation action constants
const (
	NavActionNext       = "next"
	NavActionPrev       = "prev"
	NavActionBack       = "back"
	NavActionForward    = "forward"
	NavActionToggleHelp = "toggle_help"
)
