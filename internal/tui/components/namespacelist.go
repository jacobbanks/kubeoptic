package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/models"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/styles"
)

// namespaceItem represents a namespace item in the list
type namespaceItem struct {
	name   string
	status string
}

// FilterValue implements list.Item
func (n namespaceItem) FilterValue() string {
	return n.name
}

// Title implements list.Item
func (n namespaceItem) Title() string {
	return n.name
}

// Description implements list.Item
func (n namespaceItem) Description() string {
	return fmt.Sprintf("Status: %s", n.status)
}

// namespaceDelegate handles rendering of namespace items
type namespaceDelegate struct {
	theme styles.Theme
}

// Height implements list.ItemDelegate
func (d namespaceDelegate) Height() int {
	return 1
}

// Spacing implements list.ItemDelegate
func (d namespaceDelegate) Spacing() int {
	return 0
}

// Update implements list.ItemDelegate
func (d namespaceDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render implements list.ItemDelegate
func (d namespaceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item := listItem.(namespaceItem)

	// Determine colors based on status
	var nameColor, statusColor lipgloss.Color
	switch item.status {
	case "Active":
		nameColor = styles.StatusRunning
		statusColor = styles.StatusRunning
	case "Terminating":
		nameColor = styles.StatusTerminated
		statusColor = styles.StatusTerminated
	default:
		nameColor = styles.Gray
		statusColor = styles.StatusUnknown
	}

	// Handle selection state
	if index == m.Index() {
		nameColor = d.theme.Highlight
		statusColor = d.theme.Highlight
	}

	// Create styled content
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).Faint(true)

	nameText := nameStyle.Render(item.name)
	statusText := statusStyle.Render(fmt.Sprintf(" (%s)", item.status))

	// Render with proper spacing
	content := nameText + statusText
	if index == m.Index() {
		content = lipgloss.NewStyle().
			Background(styles.Selection).
			Foreground(d.theme.Highlight).
			Render(" " + content + " ")
	}

	fmt.Fprint(w, content)
}

// NamespaceList represents the namespace list component
type NamespaceList struct {
	list      list.Model
	theme     styles.Theme
	kubeoptic *models.Kubeoptic
	focused   bool
	width     int
	height    int
}

// NewNamespaceList creates a new namespace list component
func NewNamespaceList(kubeoptic *models.Kubeoptic) *NamespaceList {
	theme := styles.DefaultTheme()
	delegate := namespaceDelegate{theme: theme}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Namespaces"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Background(theme.Primary).
		Foreground(styles.White).
		Padding(0, 1)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(theme.Primary)
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(styles.Gray)

	return &NamespaceList{
		list:      l,
		theme:     theme,
		kubeoptic: kubeoptic,
		focused:   false,
	}
}

// LoadNamespaces loads namespaces and updates the list
func (nl *NamespaceList) LoadNamespaces() tea.Cmd {
	namespaces := nl.kubeoptic.GetNamespaces()
	items := make([]list.Item, len(namespaces))

	for i, ns := range namespaces {
		items[i] = namespaceItem{
			name:   ns.Name,
			status: string(ns.Status),
		}
	}

	cmd := nl.list.SetItems(items)
	return cmd
}

// Init implements tea.Model
func (nl *NamespaceList) Init() tea.Cmd {
	return nl.LoadNamespaces()
}

// Update implements tea.Model
func (nl *NamespaceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		nl.width = msg.Width
		nl.height = msg.Height
		nl.list.SetWidth(msg.Width)
		nl.list.SetHeight(msg.Height - 3) // Leave space for title and status

	case tea.KeyMsg:
		if !nl.focused {
			return nl, nil
		}

		switch msg.String() {
		case "enter":
			// Handle namespace selection
			if selectedItem := nl.list.SelectedItem(); selectedItem != nil {
				item := selectedItem.(namespaceItem)
				return nl, tea.Batch(
					func() tea.Msg {
						err := nl.kubeoptic.SelectNamespace(item.name)
						if err != nil {
							return tui.ErrorMsg{
								Error:   err,
								Context: "selecting namespace",
							}
						}
						return tui.NamespacesLoadedMsg{
							Namespaces: []string{item.name},
							Error:      nil,
						}
					},
				)
			}

		case "r", "ctrl+r":
			// Refresh namespaces
			return nl, nl.LoadNamespaces()

		case "/":
			// Enable filtering
			nl.list.SetFilteringEnabled(true)
			nl.list, cmd = nl.list.Update(msg)
			return nl, cmd

		case "esc":
			// Clear filter or go back
			if nl.list.FilterState() == list.Filtering {
				nl.list.ResetFilter()
				return nl, nil
			}
			return nl, func() tea.Msg {
				return tui.NavigateMsg{Direction: tui.NavigateBack}
			}
		}

	case tui.NamespacesLoadedMsg:
		if msg.Error != nil {
			return nl, func() tea.Msg {
				return tui.ErrorMsg{
					Error:   msg.Error,
					Context: "loading namespaces",
				}
			}
		}
		return nl, nl.LoadNamespaces()

	case tui.FocusChangedMsg:
		nl.focused = msg.Focused && msg.Component == "namespacelist"
		if nl.focused {
			nl.list.Styles.Title = nl.list.Styles.Title.
				Background(nl.theme.Primary).
				Foreground(styles.White)
		} else {
			nl.list.Styles.Title = nl.list.Styles.Title.
				Background(styles.BorderInactive).
				Foreground(styles.Gray)
		}
	}

	nl.list, cmd = nl.list.Update(msg)
	return nl, cmd
}

// View implements tea.Model
func (nl *NamespaceList) View() string {
	if nl.width == 0 || nl.height == 0 {
		return "Loading namespaces..."
	}

	// Create the border style based on focus state
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderInactive)

	if nl.focused {
		borderStyle = borderStyle.BorderForeground(nl.theme.Primary)
	}

	// Add context information
	contextInfo := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Faint(true).
		Render(fmt.Sprintf("Context: %s", nl.kubeoptic.GetSelectedContext()))

	// Combine content
	content := lipgloss.JoinVertical(lipgloss.Left,
		contextInfo,
		nl.list.View(),
	)

	return borderStyle.
		Width(nl.width - 2).
		Height(nl.height - 2).
		Render(content)
}

// Focus sets focus to the component
func (nl *NamespaceList) Focus() tea.Cmd {
	nl.focused = true
	return func() tea.Msg {
		return tui.FocusChangedMsg{
			Component: "namespacelist",
			Focused:   true,
		}
	}
}

// Blur removes focus from the component
func (nl *NamespaceList) Blur() tea.Cmd {
	nl.focused = false
	return func() tea.Msg {
		return tui.FocusChangedMsg{
			Component: "namespacelist",
			Focused:   false,
		}
	}
}

// IsFocused returns whether the component is focused
func (nl *NamespaceList) IsFocused() bool {
	return nl.focused
}

// SetSize sets the component size
func (nl *NamespaceList) SetSize(width, height int) {
	nl.width = width
	nl.height = height
	nl.list.SetWidth(width)
	nl.list.SetHeight(height - 3) // Leave space for title and status
}

// GetSize returns the component size
func (nl *NamespaceList) GetSize() (int, int) {
	return nl.width, nl.height
}

// SetSearchQuery sets the search query for filtering
func (nl *NamespaceList) SetSearchQuery(query string) {
	nl.list.SetFilteringEnabled(true)
	// Note: Bubble tea list handles filtering internally
}

// GetSearchQuery gets the current search query
func (nl *NamespaceList) GetSearchQuery() string {
	return nl.list.FilterValue()
}

// ClearSearch clears the current search/filter
func (nl *NamespaceList) ClearSearch() {
	nl.list.ResetFilter()
}

// GetSearchResults returns filtered results
func (nl *NamespaceList) GetSearchResults() []interface{} {
	items := nl.list.Items()
	results := make([]interface{}, len(items))
	for i, item := range items {
		results[i] = item
	}
	return results
}

// RefreshData refreshes the namespace data
func (nl *NamespaceList) RefreshData() tea.Cmd {
	return nl.LoadNamespaces()
}

// GetLoadingState returns whether the component is loading
func (nl *NamespaceList) GetLoadingState() bool {
	return len(nl.list.Items()) == 0
}

// GetErrorState returns any error state
func (nl *NamespaceList) GetErrorState() error {
	return nil // Would implement error state tracking in a real scenario
}

// HandleKeyEvent handles key events
func (nl *NamespaceList) HandleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return nl.Update(msg)
}

// HandleMouseEvent handles mouse events
func (nl *NamespaceList) HandleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return nl, nil // Mouse support could be added here
}

// HandleWindowResize handles window resize events
func (nl *NamespaceList) HandleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return nl.Update(msg)
}

// GetStatusText returns status information
func (nl *NamespaceList) GetStatusText() string {
	count := len(nl.list.Items())
	selected := nl.list.Index() + 1
	if count == 0 {
		return "No namespaces"
	}
	return fmt.Sprintf("%d/%d namespaces", selected, count)
}

// GetStatusType returns the status type
func (nl *NamespaceList) GetStatusType() tui.StatusType {
	if len(nl.list.Items()) == 0 {
		return tui.StatusWarning
	}
	return tui.StatusInfo
}
