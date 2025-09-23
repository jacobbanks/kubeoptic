package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// ContextItem represents a Kubernetes context for the list
type ContextItem struct {
	context   services.Context
	isCurrent bool
}

// FilterValue returns the value used for filtering in the list
func (c ContextItem) FilterValue() string {
	return c.context.Name
}

// Title returns the display title for the context item
func (c ContextItem) Title() string {
	if c.isCurrent {
		return fmt.Sprintf("● %s", c.context.Name)
	}
	return c.context.Name
}

// Description returns the description for the context item
func (c ContextItem) Description() string {
	if c.isCurrent {
		return "Current context"
	}
	return ""
}

// ContextList wraps a Bubble Tea list for displaying Kubernetes contexts
type ContextList struct {
	list     list.Model
	contexts []services.Context
	current  string
}

// NewContextList creates a new context list component
func NewContextList(contexts []services.Context, currentContext string) ContextList {
	items := make([]list.Item, len(contexts))
	for i, ctx := range contexts {
		items[i] = ContextItem{
			context:   ctx,
			isCurrent: ctx.Name == currentContext,
		}
	}

	l := list.New(items, contextDelegate{}, 80, 24)
	l.Title = "Kubernetes Contexts"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowTitle(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginLeft(2)

	return ContextList{
		list:     l,
		contexts: contexts,
		current:  currentContext,
	}
}

// Init initializes the context list component
func (cl ContextList) Init() tea.Cmd {
	return nil
}

// Update handles messages for the context list
func (cl ContextList) Update(msg tea.Msg) (ContextList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Send a context selection message
			if selectedContext := cl.SelectedContext(); selectedContext != nil {
				return cl, func() tea.Msg {
					return tui.ContextSelectedMsg{Context: selectedContext}
				}
			}
		}
	}

	var cmd tea.Cmd
	cl.list, cmd = cl.list.Update(msg)
	return cl, cmd
}

// View renders the context list
func (cl ContextList) View() string {
	return cl.list.View()
}

// SelectedContext returns the currently selected context
func (cl ContextList) SelectedContext() *services.Context {
	if item := cl.list.SelectedItem(); item != nil {
		if contextItem, ok := item.(ContextItem); ok {
			return &contextItem.context
		}
	}
	return nil
}

// SetSize sets the dimensions of the context list
func (cl *ContextList) SetSize(width, height int) {
	cl.list.SetSize(width, height)
}

// SetFocus sets the focus state of the context list
func (cl *ContextList) SetFocus(focused bool) {
	if focused {
		cl.list.Styles.Title = cl.list.Styles.Title.
			Foreground(lipgloss.Color("205"))
	} else {
		cl.list.Styles.Title = cl.list.Styles.Title.
			Foreground(lipgloss.Color("240"))
	}
}

// contextDelegate defines how context items are rendered
type contextDelegate struct{}

func (d contextDelegate) Height() int                               { return 1 }
func (d contextDelegate) Spacing() int                              { return 0 }
func (d contextDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d contextDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ContextItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		// Selected item style
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
	} else {
		// Normal item style
		if i.isCurrent {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86"))
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
		}
	}

	fmt.Fprint(w, style.Render(i.Title()))
}
