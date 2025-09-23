package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
)

// ComponentRenderer defines the interface for all TUI components
// This enables parallel development of components across workstreams
type ComponentRenderer interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

// AppStateManager defines the interface for managing application state
// Used by components to access and modify application state
type AppStateManager interface {
	// View management
	GetCurrentView() models.ViewType
	SetCurrentView(models.ViewType)

	// Selection state
	GetSelectedContext() string
	SetSelectedContext(string)
	GetSelectedNamespace() string
	SetSelectedNamespace(string)
	GetSelectedPod() *services.Pod
	SetSelectedPod(*services.Pod)

	// Search state
	GetPodSearchQuery() string
	SetPodSearchQuery(string)
	GetFilteredPods() []services.Pod

	// Log state
	GetLogBuffer() []string
	IsFollowing() bool
	SetFollowing(bool)
}

// EventHandler defines interface for handling user input events
// Used for key bindings and user interactions
type EventHandler interface {
	HandleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd)
	HandleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd)
	HandleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd)
}

// DataProvider defines interface for components that provide data
// Used for components that need to fetch or update data
type DataProvider interface {
	RefreshData() tea.Cmd
	GetLoadingState() bool
	GetErrorState() error
}

// Focusable defines interface for components that can receive focus
// Used for navigation between components
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool
}

// Searchable defines interface for components that support search
// Used for components like pod lists that support filtering
type Searchable interface {
	SetSearchQuery(string)
	GetSearchQuery() string
	ClearSearch()
	GetSearchResults() []interface{}
}

// Resizable defines interface for components that need to handle window resize
// Used for responsive layout management
type Resizable interface {
	SetSize(width, height int)
	GetSize() (width, height int)
}

// StatusProvider defines interface for components that provide status information
// Used for status bar and other informational components
type StatusProvider interface {
	GetStatusText() string
	GetStatusType() StatusType
}
