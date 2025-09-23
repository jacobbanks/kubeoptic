package components

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
	"kubeoptic/internal/tui/styles"
)

const (
	// Buffer management
	maxLogLines   = 10000 // Maximum lines to keep in memory
	logChunkSize  = 100   // Process logs in chunks for better performance
	streamTimeout = 5 * time.Second

	// Performance optimizations
	maxSearchResults = 1000                   // Limit search results for performance
	debounceDelay    = 100 * time.Millisecond // Debounce search updates

	// Search
	maxSearchHistory = 50
	searchPrompt     = "Search: "
)

// LogDataProvider defines the interface for accessing log data and pod information
type LogDataProvider interface {
	GetSelectedPod() *services.Pod
	GetLogBuffer() []string
	IsFollowing() bool
	GetFocusedView() models.ViewType
	GetSelectedContext() string
	GetSelectedNamespace() string
}

// LogViewer represents the log viewing component
type LogViewer struct {
	// Core components
	viewport    viewport.Model
	searchInput textinput.Model

	// State
	dataProvider LogDataProvider
	focused      bool
	width        int
	height       int

	// Log management
	logLines      []string
	filteredLines []string
	logStream     io.ReadCloser
	streamCtx     context.Context
	streamCancel  context.CancelFunc

	// Display options
	followMode     bool
	showTimestamps bool
	wrapLines      bool

	// Search functionality
	searchMode    bool
	searchQuery   string
	searchResults []int // line numbers with matches
	currentResult int
	searchHistory []string

	// Error handling
	lastError error
	showError bool

	// Styling
	styles styles.LogViewerStyles
	theme  styles.Theme

	// Key bindings
	keyMap LogViewerKeyMap
}

// LogViewerKeyMap defines the key bindings for the log viewer
type LogViewerKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Home         key.Binding
	End          key.Binding
	ToggleFollow key.Binding
	Search       key.Binding
	NextSearch   key.Binding
	PrevSearch   key.Binding
	ClearSearch  key.Binding
	ToggleWrap   key.Binding
	ToggleTime   key.Binding
	Quit         key.Binding
}

// DefaultLogViewerKeyMap returns the default key bindings
func DefaultLogViewerKeyMap() LogViewerKeyMap {
	return LogViewerKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to bottom"),
		),
		ToggleFollow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle follow"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		NextSearch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		PrevSearch: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev match"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear search"),
		),
		ToggleWrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle wrap"),
		),
		ToggleTime: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle timestamps"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// NewLogViewer creates a new log viewer component
func NewLogViewer(dataProvider LogDataProvider, width, height int) *LogViewer {
	// Initialize viewport
	vp := viewport.New(width-2, height-4) // Account for borders and search
	vp.SetContent("No logs available")

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search logs..."
	searchInput.CharLimit = 256

	// Create context for stream management
	ctx, cancel := context.WithCancel(context.Background())

	theme := styles.DefaultTheme()

	return &LogViewer{
		viewport:       vp,
		searchInput:    searchInput,
		dataProvider:   dataProvider,
		width:          width,
		height:         height,
		followMode:     true,
		showTimestamps: false,
		wrapLines:      true,
		streamCtx:      ctx,
		streamCancel:   cancel,
		styles:         styles.NewLogViewerStyles(theme, width, height, false),
		theme:          theme,
		keyMap:         DefaultLogViewerKeyMap(),
		logLines:       make([]string, 0, maxLogLines),
		filteredLines:  make([]string, 0, maxLogLines),
		searchHistory:  make([]string, 0, maxSearchHistory),
	}
}

// SetSize updates the component size
func (lv *LogViewer) SetSize(width, height int) {
	lv.width = width
	lv.height = height
	lv.viewport.Width = width - 2
	lv.viewport.Height = height - 4
	lv.searchInput.Width = width - len(searchPrompt) - 4

	// Update styles
	lv.styles = styles.NewLogViewerStyles(lv.theme, width, height, lv.focused)
}

// Focus sets the component as focused
func (lv *LogViewer) Focus() tea.Cmd {
	lv.focused = true
	lv.styles = styles.NewLogViewerStyles(lv.theme, lv.width, lv.height, true)
	return nil
}

// Blur removes focus from the component
func (lv *LogViewer) Blur() tea.Cmd {
	lv.focused = false
	lv.styles = styles.NewLogViewerStyles(lv.theme, lv.width, lv.height, false)
	lv.searchMode = false
	lv.searchInput.Blur()
	return nil
}

// IsFocused returns whether the component is focused
func (lv *LogViewer) IsFocused() bool {
	return lv.focused
}

// Init initializes the log viewer
func (lv *LogViewer) Init() tea.Cmd {
	return tea.Batch(
		lv.startLogStream(),
		lv.streamLogs(),
	)
}

// Update handles all events for the log viewer
func (lv *LogViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if lv.searchMode {
			return lv.handleSearchMode(msg)
		}
		return lv.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		lv.SetSize(msg.Width, msg.Height)

	case tui.LogChunkMsg:
		return lv.handleLogChunk(msg)

	case tui.LogStreamStartedMsg:
		lv.clearError()

	case tui.LogStreamStoppedMsg:
		lv.stopLogStream()

	case tui.ErrorMsg:
		lv.setError(msg.Error)

	case tui.ToggleFollowMsg:
		lv.followMode = !lv.followMode
		if lv.followMode {
			lv.scrollToBottom()
		}
	}

	// Update viewport
	var cmd tea.Cmd
	lv.viewport, cmd = lv.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update search input if in search mode
	if lv.searchMode {
		lv.searchInput, cmd = lv.searchInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return lv, tea.Batch(cmds...)
}

// View renders the log viewer
func (lv *LogViewer) View() string {
	// Build the main content area
	content := lv.renderLogContent()
	lv.viewport.SetContent(content)

	// Build the complete view
	var sections []string

	// Header with pod info and status
	sections = append(sections, lv.renderHeader())

	// Main viewport
	sections = append(sections, lv.styles.Container.Render(lv.viewport.View()))

	// Search bar (if active)
	if lv.searchMode {
		sections = append(sections, lv.renderSearchBar())
	}

	// Status/help bar
	sections = append(sections, lv.renderStatusBar())

	// Error message (if any)
	if lv.showError && lv.lastError != nil {
		sections = append(sections, lv.renderError())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Helper methods for handling different message types and rendering

func (lv *LogViewer) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, lv.keyMap.Search):
		lv.enterSearchMode()
		return lv, nil

	case key.Matches(msg, lv.keyMap.ToggleFollow):
		return lv, func() tea.Msg { return tui.ToggleFollowMsg{} }

	case key.Matches(msg, lv.keyMap.ToggleWrap):
		lv.wrapLines = !lv.wrapLines
		return lv, nil

	case key.Matches(msg, lv.keyMap.ToggleTime):
		lv.showTimestamps = !lv.showTimestamps
		return lv, nil

	case key.Matches(msg, lv.keyMap.NextSearch) && len(lv.searchResults) > 0:
		lv.nextSearchResult()
		return lv, nil

	case key.Matches(msg, lv.keyMap.PrevSearch) && len(lv.searchResults) > 0:
		lv.prevSearchResult()
		return lv, nil

	case key.Matches(msg, lv.keyMap.ClearSearch):
		lv.clearSearch()
		return lv, nil

	case key.Matches(msg, lv.keyMap.Home):
		lv.viewport.GotoTop()
		return lv, nil

	case key.Matches(msg, lv.keyMap.End):
		lv.viewport.GotoBottom()
		return lv, nil
	}

	return lv, nil
}

func (lv *LogViewer) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		lv.executeSearch()
		return lv, nil

	case tea.KeyEsc:
		lv.exitSearchMode()
		return lv, nil
	}

	// Let search input handle the key
	var cmd tea.Cmd
	lv.searchInput, cmd = lv.searchInput.Update(msg)

	// Update search results as user types
	if lv.searchInput.Value() != lv.searchQuery {
		lv.searchQuery = lv.searchInput.Value()
		lv.updateSearchResults()
	}

	return lv, cmd
}

func (lv *LogViewer) handleLogChunk(msg tui.LogChunkMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		lv.setError(msg.Error)
		return lv, lv.streamLogs() // Continue trying to read
	}

	if msg.EOF {
		return lv, nil // Stream ended
	}

	// Add new log data
	lv.appendLogData(msg.Data)

	// Auto-scroll if in follow mode
	if lv.followMode {
		lv.scrollToBottom()
	}

	// Continue reading
	return lv, lv.streamLogs()
}

// StartLogStream initiates log streaming for the selected pod
func (lv *LogViewer) startLogStream() tea.Cmd {
	return func() tea.Msg {
		pod := lv.dataProvider.GetSelectedPod()
		if pod == nil {
			return tui.ErrorMsg{
				Error:   fmt.Errorf("no pod selected"),
				Context: "log streaming",
			}
		}

		return tui.LogStreamStartedMsg{Pod: pod}
	}
}

// streamLogs reads from the log stream asynchronously
func (lv *LogViewer) streamLogs() tea.Cmd {
	return func() tea.Msg {
		if lv.logStream == nil {
			// Try to get the stream from dataProvider
			buffer := lv.dataProvider.GetLogBuffer()
			if len(buffer) > 0 {
				// Process existing buffer
				data := strings.Join(buffer, "\n")
				return tui.LogChunkMsg{Data: data}
			}
			return nil
		}

		// Read from stream with timeout
		reader := bufio.NewReader(lv.logStream)

		// Set read deadline
		if closer, ok := lv.logStream.(interface{ SetReadDeadline(time.Time) error }); ok {
			closer.SetReadDeadline(time.Now().Add(streamTimeout))
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return tui.LogChunkMsg{EOF: true}
			}
			return tui.LogChunkMsg{Error: err}
		}

		return tui.LogChunkMsg{Data: strings.TrimSuffix(line, "\n")}
	}
}

// Log management methods
func (lv *LogViewer) appendLogData(data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Add to log buffer
		lv.logLines = append(lv.logLines, line)

		// Trim buffer if too large
		if len(lv.logLines) > maxLogLines {
			lv.logLines = lv.logLines[len(lv.logLines)-maxLogLines:]
		}
	}

	// Update filtered lines
	lv.updateFilteredLines()
}

// updateFilteredLines updates the filtered lines based on search query with performance optimizations
func (lv *LogViewer) updateFilteredLines() {
	if lv.searchQuery == "" {
		lv.filteredLines = lv.logLines
		lv.searchResults = make([]int, 0)
		return
	}

	// Performance optimization: limit search results and use case-insensitive search
	lv.filteredLines = make([]string, 0, len(lv.logLines))
	lv.searchResults = make([]int, 0, maxSearchResults)

	searchLower := strings.ToLower(lv.searchQuery)

	for i, line := range lv.logLines {
		if strings.Contains(strings.ToLower(line), searchLower) {
			lv.filteredLines = append(lv.filteredLines, line)

			// Limit search results for performance
			if len(lv.searchResults) < maxSearchResults {
				lv.searchResults = append(lv.searchResults, i)
			}
		}
	}

	// If we hit the limit, indicate there are more results
	if len(lv.searchResults) >= maxSearchResults {
		// Could add a "more results available" indicator here
	}
}

// Search functionality

// Search functionality
func (lv *LogViewer) enterSearchMode() {
	lv.searchMode = true
	lv.searchInput.Focus()
}

func (lv *LogViewer) exitSearchMode() {
	lv.searchMode = false
	lv.searchInput.Blur()
}

func (lv *LogViewer) executeSearch() {
	query := lv.searchInput.Value()
	if query != "" {
		lv.searchQuery = query
		lv.addToSearchHistory(query)
		lv.updateSearchResults()
	}
	lv.exitSearchMode()
}

func (lv *LogViewer) updateSearchResults() {
	lv.updateFilteredLines()
	if len(lv.searchResults) > 0 {
		lv.currentResult = 0
		lv.jumpToSearchResult(lv.currentResult)
	}
}

// Search functionality

func (lv *LogViewer) nextSearchResult() {
	if len(lv.searchResults) == 0 {
		return
	}
	lv.currentResult = (lv.currentResult + 1) % len(lv.searchResults)
	lv.jumpToSearchResult(lv.currentResult)
}

func (lv *LogViewer) prevSearchResult() {
	if len(lv.searchResults) == 0 {
		return
	}
	lv.currentResult = (lv.currentResult - 1 + len(lv.searchResults)) % len(lv.searchResults)
	lv.jumpToSearchResult(lv.currentResult)
}

func (lv *LogViewer) jumpToSearchResult(index int) {
	if index < 0 || index >= len(lv.searchResults) {
		return
	}

	lineNum := lv.searchResults[index]
	// Calculate viewport position to center the line
	viewportHeight := lv.viewport.Height
	targetLine := lineNum - viewportHeight/2
	if targetLine < 0 {
		targetLine = 0
	}

	// Set viewport to show the target line
	lv.viewport.SetYOffset(targetLine)
}

func (lv *LogViewer) clearSearch() {
	lv.searchQuery = ""
	lv.searchResults = make([]int, 0)
	lv.currentResult = 0
	lv.searchInput.SetValue("")
	lv.updateFilteredLines()
}

func (lv *LogViewer) addToSearchHistory(query string) {
	// Add to history if not already present
	for _, existing := range lv.searchHistory {
		if existing == query {
			return
		}
	}

	lv.searchHistory = append(lv.searchHistory, query)
	if len(lv.searchHistory) > maxSearchHistory {
		lv.searchHistory = lv.searchHistory[1:]
	}
}

// Utility methods
func (lv *LogViewer) scrollToBottom() {
	lv.viewport.GotoBottom()
}

func (lv *LogViewer) setError(err error) {
	lv.lastError = err
	lv.showError = true
}

func (lv *LogViewer) clearError() {
	lv.lastError = nil
	lv.showError = false
}

func (lv *LogViewer) stopLogStream() {
	if lv.logStream != nil {
		lv.logStream.Close()
		lv.logStream = nil
	}
	if lv.streamCancel != nil {
		lv.streamCancel()
	}
}

// Rendering methods
func (lv *LogViewer) renderLogContent() string {
	if len(lv.filteredLines) == 0 {
		return lv.styles.EmptyState.Render("No logs available")
	}

	var lines []string
	for i, line := range lv.filteredLines {
		rendered := lv.renderLogLine(line, i)
		lines = append(lines, rendered)
	}

	return strings.Join(lines, "\n")
}

func (lv *LogViewer) renderLogLine(line string, index int) string {
	// Apply syntax highlighting based on log level
	style := lv.styles.LogLine

	lowerLine := strings.ToLower(line)
	switch {
	case strings.Contains(lowerLine, "error"):
		style = lv.styles.ErrorLog
	case strings.Contains(lowerLine, "warn"):
		style = lv.styles.WarningLog
	case strings.Contains(lowerLine, "info"):
		style = lv.styles.InfoLog
	case strings.Contains(lowerLine, "debug"):
		style = lv.styles.DebugLog
	}

	// Highlight search matches
	if lv.searchQuery != "" && strings.Contains(strings.ToLower(line), strings.ToLower(lv.searchQuery)) {
		// Simple highlighting - in a full implementation, you'd want proper regex highlighting
		line = strings.ReplaceAll(line, lv.searchQuery,
			lipgloss.NewStyle().Background(lv.theme.Highlight).Render(lv.searchQuery))
	}

	return style.Render(line)
}

func (lv *LogViewer) renderHeader() string {
	pod := lv.dataProvider.GetSelectedPod()
	if pod == nil {
		return lv.styles.Title.Render("Log Viewer - No Pod Selected")
	}

	title := fmt.Sprintf("Logs: %s/%s", pod.Namespace, pod.Name)
	if lv.followMode {
		title += " [FOLLOW]"
	}

	return lv.styles.Title.Render(title)
}

func (lv *LogViewer) renderSearchBar() string {
	prompt := lv.styles.Title.Render(searchPrompt)
	input := lv.searchInput.View()

	if len(lv.searchResults) > 0 {
		status := fmt.Sprintf(" (%d/%d)", lv.currentResult+1, len(lv.searchResults))
		input += lv.styles.Title.Render(status)
	}

	return prompt + input
}

func (lv *LogViewer) renderStatusBar() string {
	var status []string

	// Follow mode indicator
	if lv.followMode {
		status = append(status, "FOLLOW")
	}

	// Line count
	status = append(status, fmt.Sprintf("%d lines", len(lv.filteredLines)))

	// Search status
	if lv.searchQuery != "" {
		status = append(status, fmt.Sprintf("Search: %s", lv.searchQuery))
	}

	// Key hints
	if !lv.searchMode {
		status = append(status, "/ search • f follow • q quit")
	}

	return lv.styles.Title.Render(strings.Join(status, " | "))
}

func (lv *LogViewer) renderError() string {
	if lv.lastError == nil {
		return ""
	}

	return styles.ErrorMessageStyles(lv.theme, lv.width).Render(
		fmt.Sprintf("Error: %v", lv.lastError))
}
