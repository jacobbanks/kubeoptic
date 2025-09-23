package events

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/keys"
	"kubeoptic/internal/tui/navigation"
)

// EventRouter handles event routing and delegation for the TUI application
type EventRouter struct {
	keyMap     keys.KeyMap
	navigation *navigation.NavigationState
}

// NewEventRouter creates a new event router
func NewEventRouter(nav *navigation.NavigationState) *EventRouter {
	return &EventRouter{
		keyMap:     keys.DefaultKeyMap(),
		navigation: nav,
	}
}

// RouteEvent routes events to the appropriate handlers based on current state
func (er *EventRouter) RouteEvent(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return er.handleKeyEvent(msg)
	case tea.MouseMsg:
		return er.handleMouseEvent(msg)
	case tea.WindowSizeMsg:
		return er.handleWindowResize(msg)
	default:
		return nil, nil, false
	}
}

// handleKeyEvent processes keyboard input
func (er *EventRouter) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	// First check for global keys that work in any context
	if handled, cmd := keys.HandleGlobalKeys(msg, er.keyMap); handled {
		return nil, cmd, true
	}

	// Handle navigation keys
	if navCmd := er.navigation.HandleNavigation(msg, er.keyMap); navCmd != nil {
		return nil, navCmd, true
	}

	// Check if we're in search mode
	if er.navigation.IsSearchMode() {
		return er.handleSearchKeys(msg)
	}

	// Route to view-specific handlers based on current view
	switch er.navigation.GetCurrentView() {
	case models.ContextView:
		return er.handleContextViewKeys(msg)
	case models.NamespaceView:
		return er.handleNamespaceViewKeys(msg)
	case models.PodView:
		return er.handlePodViewKeys(msg)
	case models.LogView:
		return er.handleLogViewKeys(msg)
	}

	return nil, nil, false
}

// handleContextViewKeys handles keys specific to the context view
func (er *EventRouter) handleContextViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, er.keyMap.Enter):
		// Select context and navigate to namespace view
		return nil, tea.Batch(
			func() tea.Msg { return tui.ContextSelectedMsg{} },
			er.navigation.NavigateToView(models.NamespaceView),
		), true

	case key.Matches(msg, er.keyMap.Refresh):
		return nil, func() tea.Msg {
			return tui.RefreshDataMsg{}
		}, true

	case keys.IsNavigationKey(msg, er.keyMap):
		// Let the component handle navigation
		return nil, nil, false

	default:
		return nil, nil, false
	}
}

// handleNamespaceViewKeys handles keys specific to the namespace view
func (er *EventRouter) handleNamespaceViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	focusedPanel := er.navigation.GetFocusedPanel()

	switch {
	case key.Matches(msg, er.keyMap.Enter):
		if focusedPanel == navigation.FocusNamespace {
			// Select namespace and navigate to pod view
			return nil, tea.Batch(
				func() tea.Msg { return tui.NamespacesLoadedMsg{} },
				er.navigation.NavigateToView(models.PodView),
			), true
		}
		return nil, nil, false

	case key.Matches(msg, er.keyMap.Refresh):
		return nil, func() tea.Msg {
			return tui.RefreshDataMsg{}
		}, true

	case keys.IsNavigationKey(msg, er.keyMap):
		// Let the component handle navigation
		return nil, nil, false

	default:
		return nil, nil, false
	}
}

// handlePodViewKeys handles keys specific to the pod view
func (er *EventRouter) handlePodViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	focusedPanel := er.navigation.GetFocusedPanel()

	switch {
	case key.Matches(msg, er.keyMap.Enter):
		if focusedPanel == navigation.FocusPod {
			// Select pod and navigate to log view
			return nil, tea.Batch(
				func() tea.Msg { return tui.PodSelectedMsg{} },
				er.navigation.NavigateToView(models.LogView),
			), true
		}
		return nil, nil, false

	case key.Matches(msg, er.keyMap.Refresh):
		return nil, func() tea.Msg {
			return tui.RefreshDataMsg{}
		}, true

	case key.Matches(msg, er.keyMap.ClearSearch):
		if er.navigation.IsSearchMode() {
			return nil, er.navigation.ExitSearchMode(), true
		}
		return nil, func() tea.Msg {
			return tui.ClearSearchMsg{}
		}, true

	case keys.IsNavigationKey(msg, er.keyMap):
		// Let the component handle navigation
		return nil, nil, false

	default:
		return nil, nil, false
	}
}

// handleLogViewKeys handles keys specific to the log view
func (er *EventRouter) handleLogViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, er.keyMap.Follow):
		return nil, func() tea.Msg {
			return tui.ToggleFollowMsg{}
		}, true

	case key.Matches(msg, er.keyMap.SaveLogs):
		return nil, func() tea.Msg {
			return keys.SaveLogsMsg{}
		}, true

	case key.Matches(msg, er.keyMap.ToggleWrap):
		return nil, func() tea.Msg {
			return ToggleWrapMsg{}
		}, true

	case key.Matches(msg, er.keyMap.ScrollUp, er.keyMap.ScrollDown,
		er.keyMap.PageUp, er.keyMap.PageDown, er.keyMap.Home, er.keyMap.End):
		// Let the log viewer component handle scrolling
		return nil, nil, false

	default:
		return nil, nil, false
	}
}

// handleSearchKeys handles keys when in search mode
func (er *EventRouter) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, er.keyMap.Back):
		return nil, er.navigation.ExitSearchMode(), true

	case key.Matches(msg, er.keyMap.Enter):
		// Apply search and exit search mode
		return nil, er.navigation.ExitSearchMode(), true

	case key.Matches(msg, er.keyMap.ClearSearch):
		return nil, tea.Batch(
			func() tea.Msg { return tui.ClearSearchMsg{} },
			er.navigation.ExitSearchMode(),
		), true

	default:
		// Let the search input component handle typing
		return nil, nil, false
	}
}

// handleMouseEvent processes mouse input
func (er *EventRouter) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.Type {
	case tea.MouseLeft:
		// Handle click events - can be used for focus management
		return nil, func() tea.Msg {
			return MouseClickMsg{
				X:      msg.X,
				Y:      msg.Y,
				Button: "left",
			}
		}, true

	case tea.MouseWheelUp, tea.MouseWheelDown:
		// Handle scroll events in log view
		if er.navigation.GetCurrentView() == models.LogView {
			direction := "up"
			if msg.Type == tea.MouseWheelDown {
				direction = "down"
			}
			return nil, func() tea.Msg {
				return MouseScrollMsg{
					Direction: direction,
					X:         msg.X,
					Y:         msg.Y,
				}
			}, true
		}

	default:
		return nil, nil, false
	}

	return nil, nil, false
}

// handleWindowResize processes window resize events
func (er *EventRouter) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd, bool) {
	return nil, func() tea.Msg {
		return tui.WindowResizeMsg{
			Width:  msg.Width,
			Height: msg.Height,
		}
	}, true
}

// EventHandler implements the EventHandler interface from the TUI interfaces
type EventHandler struct {
	router *EventRouter
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(nav *navigation.NavigationState) *EventHandler {
	return &EventHandler{
		router: NewEventRouter(nav),
	}
}

// HandleKeyEvent implements the EventHandler interface
func (eh *EventHandler) HandleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	model, cmd, _ := eh.router.handleKeyEvent(msg)
	return model, cmd
}

// HandleMouseEvent implements the EventHandler interface
func (eh *EventHandler) HandleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	model, cmd, _ := eh.router.handleMouseEvent(msg)
	return model, cmd
}

// HandleWindowResize implements the EventHandler interface
func (eh *EventHandler) HandleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	model, cmd, _ := eh.router.handleWindowResize(msg)
	return model, cmd
}

// Custom message types for event handling
type MouseClickMsg struct {
	X, Y   int
	Button string
}

type MouseScrollMsg struct {
	Direction string
	X, Y      int
}

type ToggleWrapMsg struct{}

type FocusComponentMsg struct {
	Component string
}

// EventProcessor provides high-level event processing for the main app
type EventProcessor struct {
	eventHandler *EventHandler
	navigation   *navigation.NavigationState
}

// NewEventProcessor creates a new event processor
func NewEventProcessor() *EventProcessor {
	nav := navigation.NewNavigationState()
	return &EventProcessor{
		eventHandler: NewEventHandler(nav),
		navigation:   nav,
	}
}

// ProcessEvent processes an event and returns appropriate commands
func (ep *EventProcessor) ProcessEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route the event through the event router
	model, cmd, handled := ep.eventHandler.router.RouteEvent(msg)

	if handled {
		return model, cmd
	}

	// If not handled by router, let components handle it
	return nil, nil
}

// GetNavigation returns the navigation state
func (ep *EventProcessor) GetNavigation() *navigation.NavigationState {
	return ep.navigation
}

// GetKeyMap returns the current key map
func (ep *EventProcessor) GetKeyMap() keys.KeyMap {
	return ep.eventHandler.router.keyMap
}

// UpdateKeyMap updates the key map (for customization)
func (ep *EventProcessor) UpdateKeyMap(keyMap keys.KeyMap) {
	ep.eventHandler.router.keyMap = keyMap
}

// ComponentEventDelegate delegates events to specific components
type ComponentEventDelegate struct {
	componentType string
	focusState    navigation.FocusState
	keyMap        keys.KeyMap
}

// NewComponentEventDelegate creates a new component event delegate
func NewComponentEventDelegate(componentType string, focusState navigation.FocusState, keyMap keys.KeyMap) *ComponentEventDelegate {
	return &ComponentEventDelegate{
		componentType: componentType,
		focusState:    focusState,
		keyMap:        keyMap,
	}
}

// ShouldHandleEvent determines if this component should handle the event
func (ced *ComponentEventDelegate) ShouldHandleEvent(currentFocus navigation.FocusState) bool {
	return ced.focusState == currentFocus
}

// DelegateKeyEvent delegates a key event to the component
func (ced *ComponentEventDelegate) DelegateKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	// This method can be implemented by specific component delegates
	// For now, return false to indicate the event wasn't handled
	return nil, nil, false
}
