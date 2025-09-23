package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// PodItem represents a pod in the list with required interfaces
type PodItem struct {
	Pod services.Pod
}

// FilterValue implements list.Item interface for filtering
func (p PodItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s", p.Pod.Name, p.Pod.Namespace, p.Pod.Status)
}

// Title implements list.DefaultItem interface
func (p PodItem) Title() string {
	return p.Pod.Name
}

// Description implements list.DefaultItem interface
func (p PodItem) Description() string {
	return fmt.Sprintf("Status: %s | Namespace: %s", p.Pod.Status, p.Pod.Namespace)
}

// PodList manages the pod list component
type PodList struct {
	list      list.Model
	pods      []services.Pod
	focused   bool
	width     int
	height    int
	searching bool
}

// NewPodList creates a new pod list component
func NewPodList(pods []services.Pod, width, height int) *PodList {
	items := make([]list.Item, len(pods))
	for i, pod := range pods {
		items[i] = PodItem{Pod: pod}
	}

	// Create custom delegate for pod-specific styling
	delegate := newPodDelegate()

	l := list.New(items, delegate, width, height)
	l.Title = "Pods"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	return &PodList{
		list:   l,
		pods:   pods,
		width:  width,
		height: height,
	}
}

// Init implements tea.Model interface
func (p *PodList) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface
func (p *PodList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.list.SetSize(p.width, p.height)
		return p, nil

	case tea.KeyMsg:
		// Handle special keys first
		switch msg.String() {
		case "enter":
			if selectedItem := p.list.SelectedItem(); selectedItem != nil {
				if podItem, ok := selectedItem.(PodItem); ok {
					// Return a PodSelectedMsg for the parent to handle
					return p, func() tea.Msg {
						return tui.PodSelectedMsg{Pod: &podItem.Pod}
					}
				}
			}
			return p, nil

		case "/":
			p.searching = true
			p.list.SetFilteringEnabled(true)
			// The list will automatically enter filtering mode when the user types
			return p, nil

		case "esc":
			if p.list.FilterState() == list.Filtering {
				p.searching = false
				p.list.ResetFilter()
			}
			return p, nil
		}

	case tui.PodsLoadedMsg:
		if msg.Error != nil {
			// Handle error - could add error state to component
			return p, nil
		}
		p.UpdatePods(msg.Pods)
		return p, nil

	case tui.SearchQueryChangedMsg:
		// Update search query - let the list handle filtering automatically
		p.list.SetFilteringEnabled(true)
		return p, nil
	}

	// Update the list
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

// View implements tea.Model interface
func (p *PodList) View() string {
	if !p.focused {
		// Add a subtle style for unfocused state
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(p.list.View())
	}

	// Focused state with highlighted border
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Render(p.list.View())
}

// UpdatePods updates the pod list with new data
func (p *PodList) UpdatePods(pods []services.Pod) {
	p.pods = pods
	items := make([]list.Item, len(pods))
	for i, pod := range pods {
		items[i] = PodItem{Pod: pod}
	}
	p.list.SetItems(items)
}

// GetSelectedPod returns the currently selected pod
func (p *PodList) GetSelectedPod() *services.Pod {
	if selectedItem := p.list.SelectedItem(); selectedItem != nil {
		if podItem, ok := selectedItem.(PodItem); ok {
			return &podItem.Pod
		}
	}
	return nil
}

// Focus sets the component as focused
func (p *PodList) Focus() tea.Cmd {
	p.focused = true
	return nil
}

// Blur removes focus from the component
func (p *PodList) Blur() {
	p.focused = false
}

// IsFocused returns whether the component is focused
func (p *PodList) IsFocused() bool {
	return p.focused
}

// SetSize updates the component size
func (p *PodList) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.list.SetSize(width, height)
}

// GetSize returns the current component size
func (p *PodList) GetSize() (int, int) {
	return p.width, p.height
}

// SetSearchQuery sets the search query for filtering
func (p *PodList) SetSearchQuery(query string) {
	// This would trigger filtering - Bubble Tea list handles this internally
	p.searching = query != ""
}

// GetSearchQuery returns the current search query
func (p *PodList) GetSearchQuery() string {
	if p.list.FilterState() == list.Filtering {
		return p.list.FilterValue()
	}
	return ""
}

// ClearSearch clears the current search
func (p *PodList) ClearSearch() {
	p.searching = false
	p.list.ResetFilter()
}

// GetSearchResults returns filtered results (handled by bubbles list internally)
func (p *PodList) GetSearchResults() []interface{} {
	results := make([]interface{}, 0)
	for _, item := range p.list.VisibleItems() {
		results = append(results, item)
	}
	return results
}

// podDelegate creates a custom delegate for pod list items with status color coding
type podDelegate struct {
	styles     podDelegateStyles
	showLabels bool
}

type podDelegateStyles struct {
	normal    lipgloss.Style
	selected  lipgloss.Style
	running   lipgloss.Style
	pending   lipgloss.Style
	failed    lipgloss.Style
	succeeded lipgloss.Style
	unknown   lipgloss.Style
}

func newPodDelegate() *podDelegate {
	return &podDelegate{
		styles: podDelegateStyles{
			normal:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
			selected:  lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true),
			running:   lipgloss.NewStyle().Foreground(lipgloss.Color("46")),  // Green
			pending:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange
			failed:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
			succeeded: lipgloss.NewStyle().Foreground(lipgloss.Color("82")),  // Light green
			unknown:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // Gray
		},
		showLabels: true,
	}
}

func (d *podDelegate) Height() int  { return 2 }
func (d *podDelegate) Spacing() int { return 1 }

func (d *podDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d *podDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	podItem, ok := listItem.(PodItem)
	if !ok {
		return
	}

	pod := podItem.Pod
	isSelected := index == m.Index()

	// Get status-specific style
	statusStyle := d.getStatusStyle(pod.Status)

	var nameStyle, descStyle lipgloss.Style
	if isSelected {
		nameStyle = d.styles.selected
		descStyle = d.styles.selected
	} else {
		nameStyle = statusStyle
		descStyle = d.styles.normal
	}

	// Format the pod name with status indicator
	statusIndicator := d.getStatusIndicator(pod.Status)
	name := fmt.Sprintf("%s %s", statusIndicator, pod.Name)

	// Format description with namespace and labels if available
	desc := fmt.Sprintf("Namespace: %s", pod.Namespace)
	if len(pod.Labels) > 0 && d.showLabels {
		labelStrs := make([]string, 0, len(pod.Labels))
		for k, v := range pod.Labels {
			labelStrs = append(labelStrs, fmt.Sprintf("%s=%s", k, v))
		}
		if len(labelStrs) > 0 {
			desc += fmt.Sprintf(" | Labels: %s", strings.Join(labelStrs[:min(2, len(labelStrs))], ", "))
			if len(labelStrs) > 2 {
				desc += "..."
			}
		}
	}

	// Render with proper width handling
	w.Write([]byte(nameStyle.Render(name) + "\n"))
	w.Write([]byte(descStyle.Render(desc)))
}

func (d *podDelegate) getStatusStyle(status services.PodStatus) lipgloss.Style {
	switch status {
	case services.PodRunning:
		return d.styles.running
	case services.PodPending:
		return d.styles.pending
	case services.PodFailed:
		return d.styles.failed
	case services.PodSucceeded:
		return d.styles.succeeded
	default:
		return d.styles.unknown
	}
}

func (d *podDelegate) getStatusIndicator(status services.PodStatus) string {
	switch status {
	case services.PodRunning:
		return "●" // Green dot
	case services.PodPending:
		return "◐" // Half circle
	case services.PodFailed:
		return "✗" // X mark
	case services.PodSucceeded:
		return "✓" // Check mark
	default:
		return "?" // Question mark
	}
}

// Helper function for min since Go doesn't have generics for this in older versions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
