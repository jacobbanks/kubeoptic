package keys

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
)

// KeyMap defines all available key bindings for the application
type KeyMap struct {
	// Global navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Back  key.Binding
	Home  key.Binding
	End   key.Binding

	// Tab navigation (between panels in main view)
	NextPanel key.Binding
	PrevPanel key.Binding

	// View-specific actions
	Search      key.Binding
	ClearSearch key.Binding
	Follow      key.Binding
	Refresh     key.Binding

	// Application controls
	Help key.Binding
	Quit key.Binding

	// Log viewer specific
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	ToggleWrap key.Binding
	SaveLogs   key.Binding

	// Context/Pod management
	SwitchContext   key.Binding
	SelectNamespace key.Binding
	SelectPod       key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "bottom"),
		),

		// Tab navigation
		NextPanel: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		PrevPanel: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),

		// View-specific actions
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "clear search"),
		),
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle follow"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),

		// Application controls
		Help: key.NewBinding(
			key.WithKeys("?", "F1"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),

		// Log viewer specific
		ScrollUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", " "),
			key.WithHelp("pgdown/space", "page down"),
		),
		ToggleWrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle wrap"),
		),
		SaveLogs: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save logs"),
		),

		// Context/Pod management
		SwitchContext: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "contexts"),
		),
		SelectNamespace: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "namespaces"),
		),
		SelectPod: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pods"),
		),
	}
}

// ViewKeyMap returns context-sensitive key bindings based on the current view
type ViewKeyMap struct {
	Global KeyMap
	View   []key.Binding
	Help   []key.Binding
}

// GetViewKeyMap returns the appropriate key map for the given view
func GetViewKeyMap(view models.ViewType, keyMap KeyMap) ViewKeyMap {
	switch view {
	case models.ContextView:
		return ViewKeyMap{
			Global: keyMap,
			View: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back,
				keyMap.Search, keyMap.Refresh, keyMap.NextPanel,
			},
			Help: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Search,
				keyMap.Refresh, keyMap.Help, keyMap.Quit,
			},
		}

	case models.NamespaceView:
		return ViewKeyMap{
			Global: keyMap,
			View: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back,
				keyMap.Search, keyMap.Refresh, keyMap.NextPanel,
				keyMap.SwitchContext,
			},
			Help: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back,
				keyMap.Search, keyMap.SwitchContext, keyMap.Help, keyMap.Quit,
			},
		}

	case models.PodView:
		return ViewKeyMap{
			Global: keyMap,
			View: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back,
				keyMap.Search, keyMap.ClearSearch, keyMap.Refresh,
				keyMap.NextPanel, keyMap.SwitchContext, keyMap.SelectNamespace,
			},
			Help: []key.Binding{
				keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back,
				keyMap.Search, keyMap.Refresh, keyMap.SwitchContext,
				keyMap.SelectNamespace, keyMap.Help, keyMap.Quit,
			},
		}

	case models.LogView:
		return ViewKeyMap{
			Global: keyMap,
			View: []key.Binding{
				keyMap.ScrollUp, keyMap.ScrollDown, keyMap.PageUp, keyMap.PageDown,
				keyMap.Back, keyMap.Follow, keyMap.Search, keyMap.ToggleWrap,
				keyMap.SaveLogs, keyMap.Refresh, keyMap.Home, keyMap.End,
			},
			Help: []key.Binding{
				keyMap.ScrollUp, keyMap.ScrollDown, keyMap.PageUp, keyMap.PageDown,
				keyMap.Back, keyMap.Follow, keyMap.Search, keyMap.ToggleWrap,
				keyMap.Help, keyMap.Quit,
			},
		}

	default:
		return ViewKeyMap{
			Global: keyMap,
			View:   []key.Binding{keyMap.Help, keyMap.Quit},
			Help:   []key.Binding{keyMap.Help, keyMap.Quit},
		}
	}
}

// HandleGlobalKeys handles keys that should work in any view
func HandleGlobalKeys(msg tea.KeyMsg, keyMap KeyMap) (bool, tea.Cmd) {
	switch {
	case key.Matches(msg, keyMap.Quit):
		return true, tea.Quit

	case key.Matches(msg, keyMap.Help):
		// Return a help message that the app can handle
		return true, func() tea.Msg {
			return ShowHelpMsg{}
		}

	default:
		return false, nil
	}
}

// Custom message types for key handling
type ShowHelpMsg struct{}
type HideHelpMsg struct{}
type NavigateToViewMsg struct {
	View models.ViewType
}
type ToggleFollowMsg struct{}
type RefreshDataMsg struct{}
type SaveLogsMsg struct{}

// IsNavigationKey checks if the key is a navigation key
func IsNavigationKey(msg tea.KeyMsg, keyMap KeyMap) bool {
	return key.Matches(msg, keyMap.Up) ||
		key.Matches(msg, keyMap.Down) ||
		key.Matches(msg, keyMap.Left) ||
		key.Matches(msg, keyMap.Right) ||
		key.Matches(msg, keyMap.NextPanel) ||
		key.Matches(msg, keyMap.PrevPanel)
}

// IsActionKey checks if the key is an action key (enter, back, etc.)
func IsActionKey(msg tea.KeyMsg, keyMap KeyMap) bool {
	return key.Matches(msg, keyMap.Enter) ||
		key.Matches(msg, keyMap.Back) ||
		key.Matches(msg, keyMap.Search) ||
		key.Matches(msg, keyMap.Refresh)
}

// GetKeyHelp returns the help text for context-sensitive keys
func GetKeyHelp(view models.ViewType, keyMap KeyMap) [][]key.Binding {
	switch view {
	case models.ContextView:
		return [][]key.Binding{
			{keyMap.Up, keyMap.Down, keyMap.Enter},
			{keyMap.Search, keyMap.Refresh, keyMap.Help, keyMap.Quit},
		}

	case models.NamespaceView:
		return [][]key.Binding{
			{keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back},
			{keyMap.Search, keyMap.SwitchContext, keyMap.Help, keyMap.Quit},
		}

	case models.PodView:
		return [][]key.Binding{
			{keyMap.Up, keyMap.Down, keyMap.Enter, keyMap.Back},
			{keyMap.Search, keyMap.ClearSearch, keyMap.Refresh},
			{keyMap.SwitchContext, keyMap.SelectNamespace, keyMap.Help, keyMap.Quit},
		}

	case models.LogView:
		return [][]key.Binding{
			{keyMap.ScrollUp, keyMap.ScrollDown, keyMap.PageUp, keyMap.PageDown},
			{keyMap.Back, keyMap.Follow, keyMap.Search, keyMap.ToggleWrap},
			{keyMap.Home, keyMap.End, keyMap.Help, keyMap.Quit},
		}

	default:
		return [][]key.Binding{
			{keyMap.Help, keyMap.Quit},
		}
	}
}
