package components

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// MockKubeoptic provides a mock implementation for testing
type MockKubeoptic struct {
	selectedPod *services.Pod
	logBuffer   []string
	following   bool
}

func (m *MockKubeoptic) GetSelectedPod() *services.Pod {
	return m.selectedPod
}

func (m *MockKubeoptic) GetLogBuffer() []string {
	return m.logBuffer
}

func (m *MockKubeoptic) IsFollowing() bool {
	return m.following
}

func (m *MockKubeoptic) GetFocusedView() models.ViewType {
	return models.LogView
}

func (m *MockKubeoptic) GetSelectedContext() string {
	return "test-context"
}

func (m *MockKubeoptic) GetSelectedNamespace() string {
	return "test-namespace"
}

func newMockKubeoptic() LogDataProvider {
	return &MockKubeoptic{
		selectedPod: &services.Pod{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Status:    services.PodRunning,
		},
		logBuffer: []string{
			"2023-01-01 10:00:00 INFO Starting application",
			"2023-01-01 10:00:01 ERROR Failed to connect to database",
			"2023-01-01 10:00:02 WARN Connection retrying",
			"2023-01-01 10:00:03 INFO Connected successfully",
		},
		following: true,
	}
}

func TestNewLogViewer(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test initial state
	if lv.dataProvider != dataProvider {
		t.Error("Expected dataProvider to be set")
	}

	if lv.width != 80 || lv.height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", lv.width, lv.height)
	}

	if !lv.followMode {
		t.Error("Expected follow mode to be enabled by default")
	}

	if lv.wrapLines != true {
		t.Error("Expected wrap lines to be enabled by default")
	}

	if lv.searchMode {
		t.Error("Expected search mode to be disabled by default")
	}

	if len(lv.logLines) != 0 {
		t.Error("Expected empty log lines initially")
	}
}

func TestLogViewerSetSize(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	lv.SetSize(100, 30)

	if lv.width != 100 || lv.height != 30 {
		t.Errorf("Expected dimensions 100x30, got %dx%d", lv.width, lv.height)
	}

	// Viewport should be adjusted for borders
	if lv.viewport.Width != 98 || lv.viewport.Height != 26 {
		t.Errorf("Expected viewport 98x26, got %dx%d", lv.viewport.Width, lv.viewport.Height)
	}
}

func TestLogViewerFocus(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test focus
	lv.Focus()
	if !lv.IsFocused() {
		t.Error("Expected component to be focused")
	}

	// Test blur
	lv.Blur()
	if lv.IsFocused() {
		t.Error("Expected component to be blurred")
	}

	if lv.searchMode {
		t.Error("Expected search mode to be disabled after blur")
	}
}

func TestLogViewerAppendLogData(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test single line
	lv.appendLogData("Test log line")
	if len(lv.logLines) != 1 {
		t.Errorf("Expected 1 log line, got %d", len(lv.logLines))
	}

	if lv.logLines[0] != "Test log line" {
		t.Errorf("Expected 'Test log line', got '%s'", lv.logLines[0])
	}

	// Test multiple lines
	lv.appendLogData("Line 1\nLine 2\nLine 3")
	if len(lv.logLines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lv.logLines))
	}

	// Test empty lines are filtered
	lv.appendLogData("Line 4\n\nLine 5")
	if len(lv.logLines) != 6 {
		t.Errorf("Expected 6 log lines, got %d", len(lv.logLines))
	}
}

func TestLogViewerSearchFunctionality(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Add test data
	lv.appendLogData("INFO Application started")
	lv.appendLogData("ERROR Database connection failed")
	lv.appendLogData("WARN Retrying connection")
	lv.appendLogData("INFO Connection established")

	// Test search
	lv.searchQuery = "ERROR"
	lv.updateSearchResults()

	if len(lv.searchResults) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(lv.searchResults))
	}

	if lv.searchResults[0] != 1 {
		t.Errorf("Expected search result at line 1, got %d", lv.searchResults[0])
	}

	// Test case-insensitive search
	lv.searchQuery = "info"
	lv.updateSearchResults()

	if len(lv.searchResults) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(lv.searchResults))
	}

	// Test search navigation
	lv.nextSearchResult()
	if lv.currentResult != 1 {
		t.Errorf("Expected current result 1, got %d", lv.currentResult)
	}

	lv.prevSearchResult()
	if lv.currentResult != 0 {
		t.Errorf("Expected current result 0, got %d", lv.currentResult)
	}

	// Test clear search
	lv.clearSearch()
	if lv.searchQuery != "" {
		t.Errorf("Expected empty search query after clear")
	}

	if len(lv.searchResults) != 0 {
		t.Errorf("Expected no search results after clear")
	}
}

func TestLogViewerKeyBindings(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)
	lv.Focus()

	tests := []struct {
		name        string
		key         string
		expectCmd   bool
		expectState func(*LogViewer) bool
	}{
		{"Search key", "/", false, func(lv *LogViewer) bool { return lv.searchMode }},
		{"Toggle follow", "f", true, nil},
		{"Toggle wrap", "w", false, func(lv *LogViewer) bool { return !lv.wrapLines }},
		{"Toggle timestamps", "t", false, func(lv *LogViewer) bool { return lv.showTimestamps }},
		{"Home", "g", false, nil},
		{"End", "G", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			lv.searchMode = false
			lv.wrapLines = true
			lv.showTimestamps = false

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			_, cmd := lv.Update(msg)

			if tt.expectCmd && cmd == nil {
				t.Errorf("Expected command for key %s", tt.key)
			}

			if !tt.expectCmd && cmd != nil && tt.key != "g" && tt.key != "G" {
				t.Errorf("Did not expect command for key %s", tt.key)
			}

			if tt.expectState != nil && !tt.expectState(lv) {
				t.Errorf("Expected state change for key %s", tt.key)
			}
		})
	}
}

func TestLogViewerSearchMode(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)
	lv.Focus()

	// Enter search mode
	lv.enterSearchMode()
	if !lv.searchMode {
		t.Error("Expected search mode to be active")
	}

	// Test search input
	lv.searchInput.SetValue("test query")
	lv.searchQuery = lv.searchInput.Value()

	if lv.searchQuery != "test query" {
		t.Errorf("Expected 'test query', got '%s'", lv.searchQuery)
	}

	// Test search history
	lv.addToSearchHistory("test query")
	if len(lv.searchHistory) != 1 {
		t.Errorf("Expected 1 search history item, got %d", len(lv.searchHistory))
	}

	// Test duplicate prevention
	lv.addToSearchHistory("test query")
	if len(lv.searchHistory) != 1 {
		t.Errorf("Expected 1 search history item after duplicate, got %d", len(lv.searchHistory))
	}

	// Exit search mode
	lv.exitSearchMode()
	if lv.searchMode {
		t.Error("Expected search mode to be inactive")
	}
}

func TestLogViewerErrorHandling(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test error setting
	testError := errors.New("test error")
	lv.setError(testError)

	if lv.lastError != testError {
		t.Error("Expected error to be set")
	}

	if !lv.showError {
		t.Error("Expected showError to be true")
	}

	// Test error clearing
	lv.clearError()
	if lv.lastError != nil {
		t.Error("Expected error to be cleared")
	}

	if lv.showError {
		t.Error("Expected showError to be false")
	}
}

func TestLogViewerLogChunkHandling(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test successful log chunk
	msg := tui.LogChunkMsg{
		Data: "New log line",
	}

	_, cmd := lv.handleLogChunk(msg)
	if cmd == nil {
		t.Error("Expected command to continue streaming")
	}

	if len(lv.logLines) != 1 {
		t.Errorf("Expected 1 log line, got %d", len(lv.logLines))
	}

	// Test error handling
	errorMsg := tui.LogChunkMsg{
		Error: errors.New("stream error"),
	}

	_, cmd = lv.handleLogChunk(errorMsg)
	if cmd == nil {
		t.Error("Expected command to continue streaming after error")
	}

	if lv.lastError == nil {
		t.Error("Expected error to be set")
	}

	// Test EOF
	eofMsg := tui.LogChunkMsg{
		EOF: true,
	}

	_, cmd = lv.handleLogChunk(eofMsg)
	if cmd != nil {
		t.Error("Expected no command after EOF")
	}
}

func TestLogViewerFollowMode(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test follow mode toggle
	initialFollow := lv.followMode
	msg := tui.ToggleFollowMsg{}
	lv.Update(msg)

	if lv.followMode == initialFollow {
		t.Error("Expected follow mode to toggle")
	}
}

func TestLogViewerBufferLimit(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Add more than maxLogLines
	for i := 0; i < maxLogLines+100; i++ {
		lv.appendLogData("Log line " + string(rune(i)))
	}

	if len(lv.logLines) > maxLogLines {
		t.Errorf("Expected log lines to be limited to %d, got %d", maxLogLines, len(lv.logLines))
	}
}

func TestLogViewerRenderMethods(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test header rendering
	header := lv.renderHeader()
	if !strings.Contains(header, "test-pod") {
		t.Error("Expected header to contain pod name")
	}

	// Test log content rendering with no logs
	content := lv.renderLogContent()
	if !strings.Contains(content, "No logs available") {
		t.Error("Expected empty state message")
	}

	// Test with logs
	lv.appendLogData("INFO Test log")
	lv.appendLogData("ERROR Test error")
	content = lv.renderLogContent()

	if !strings.Contains(content, "Test log") {
		t.Error("Expected log content to contain log lines")
	}

	// Test status bar rendering
	statusBar := lv.renderStatusBar()
	if !strings.Contains(statusBar, "lines") {
		t.Error("Expected status bar to contain line count")
	}

	// Test search bar rendering
	lv.searchMode = true
	lv.searchQuery = "test"
	lv.searchResults = []int{0}
	searchBar := lv.renderSearchBar()

	if !strings.Contains(searchBar, "Search:") {
		t.Error("Expected search bar to contain search prompt")
	}
}

func TestLogViewerLogLineStyling(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	tests := []struct {
		line     string
		contains string
	}{
		{"INFO Application started", "INFO"},
		{"ERROR Database failed", "ERROR"},
		{"WARN Connection issue", "WARN"},
		{"DEBUG Verbose logging", "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			rendered := lv.renderLogLine(tt.line, 0)
			// We can't easily test the styling, but we can ensure the line content is preserved
			if !strings.Contains(rendered, tt.contains) {
				t.Errorf("Expected rendered line to contain '%s'", tt.contains)
			}
		})
	}
}

func TestLogViewerView(t *testing.T) {
	dataProvider := newMockKubeoptic()
	lv := NewLogViewer(dataProvider, 80, 24)

	// Test basic view rendering
	view := lv.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Test with search mode
	lv.searchMode = true
	view = lv.View()
	if !strings.Contains(view, "Search:") {
		t.Error("Expected view to contain search bar when in search mode")
	}

	// Test with error
	lv.setError(errors.New("test error"))
	view = lv.View()
	if !strings.Contains(view, "Error:") {
		t.Error("Expected view to contain error message")
	}
}
