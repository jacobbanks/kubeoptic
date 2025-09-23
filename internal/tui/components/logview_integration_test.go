package components

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// MockLogStream simulates a log stream for integration testing
type MockLogStream struct {
	logs     []string
	position int
	closed   bool
}

func NewMockLogStream(logs []string) *MockLogStream {
	return &MockLogStream{
		logs:     logs,
		position: 0,
		closed:   false,
	}
}

func (m *MockLogStream) Read(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}

	if m.position >= len(m.logs) {
		return 0, io.EOF
	}

	line := m.logs[m.position] + "\n"
	m.position++

	copy(p, []byte(line))
	return len(line), nil
}

func (m *MockLogStream) Close() error {
	m.closed = true
	return nil
}

// StreamingMockKubeoptic extends MockKubeoptic with streaming capabilities
type StreamingMockKubeoptic struct {
	*MockKubeoptic
	logStream *MockLogStream
}

func newStreamingMockKubeoptic() *StreamingMockKubeoptic {
	return &StreamingMockKubeoptic{
		MockKubeoptic: &MockKubeoptic{
			selectedPod: &services.Pod{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Status:    services.PodRunning,
			},
			logBuffer: []string{},
			following: true,
		},
		logStream: NewMockLogStream([]string{
			"2023-01-01 10:00:00 INFO Application starting",
			"2023-01-01 10:00:01 INFO Loading configuration",
			"2023-01-01 10:00:02 ERROR Failed to connect to database",
			"2023-01-01 10:00:03 WARN Retrying connection",
			"2023-01-01 10:00:04 INFO Connected to database",
			"2023-01-01 10:00:05 INFO Application ready",
		}),
	}
}

func TestLogViewerIntegration_EndToEndWorkflow(t *testing.T) {
	dataProvider := newStreamingMockKubeoptic()
	lv := NewLogViewer(dataProvider, 120, 40)
	lv.Focus()

	t.Run("initial_state", func(t *testing.T) {
		// Test initial state
		if !lv.IsFocused() {
			t.Error("Expected log viewer to be focused")
		}

		if !lv.followMode {
			t.Error("Expected follow mode to be enabled by default")
		}

		view := lv.View()
		if !strings.Contains(view, "test-pod") {
			t.Error("Expected view to contain pod name")
		}
	})

	t.Run("log_streaming_simulation", func(t *testing.T) {
		// Simulate receiving log chunks
		logChunks := []string{
			"INFO Application starting",
			"ERROR Database connection failed",
			"WARN Retrying connection",
			"INFO Connection established",
		}

		for _, chunk := range logChunks {
			msg := tui.LogChunkMsg{Data: chunk}
			lv.Update(msg)
		}

		// Verify logs were added
		if len(lv.logLines) != len(logChunks) {
			t.Errorf("Expected %d log lines, got %d", len(logChunks), len(lv.logLines))
		}

		// Test log content rendering
		content := lv.renderLogContent()
		for _, chunk := range logChunks {
			if !strings.Contains(content, chunk) {
				t.Errorf("Expected log content to contain '%s'", chunk)
			}
		}
	})

	t.Run("search_integration", func(t *testing.T) {
		// Enter search mode
		lv.enterSearchMode()
		if !lv.searchMode {
			t.Error("Expected search mode to be active")
		}

		// Simulate typing search query
		lv.searchInput.SetValue("ERROR")
		lv.executeSearch()

		// Verify search results
		if len(lv.searchResults) == 0 {
			t.Error("Expected search results for 'ERROR'")
		}

		// Test search navigation
		if len(lv.searchResults) > 0 {
			initialResult := lv.currentResult
			lv.nextSearchResult()

			if len(lv.searchResults) > 1 && lv.currentResult == initialResult {
				t.Error("Expected current result to change")
			}
		}
	})

	t.Run("follow_mode_integration", func(t *testing.T) {
		// Test follow mode toggle
		initialFollow := lv.followMode
		msg := tui.ToggleFollowMsg{}
		lv.Update(msg)

		if lv.followMode == initialFollow {
			t.Error("Expected follow mode to toggle")
		}

		// Add new log and verify auto-scroll behavior
		newLogMsg := tui.LogChunkMsg{Data: "NEW LOG ENTRY"}
		lv.Update(newLogMsg)

		// In follow mode, should scroll to bottom
		if lv.followMode {
			// Verify viewport is at bottom (in a real implementation)
			// This is a simplified test since we can't easily test viewport position
		}
	})
}

func TestLogViewerIntegration_ErrorHandling(t *testing.T) {
	dataProvider := newStreamingMockKubeoptic()
	lv := NewLogViewer(dataProvider, 120, 40)

	t.Run("stream_error_recovery", func(t *testing.T) {
		// Simulate stream error
		errorMsg := tui.LogChunkMsg{
			Error: io.ErrUnexpectedEOF,
		}

		_, cmd := lv.Update(errorMsg)

		// Should set error state
		if lv.lastError == nil {
			t.Error("Expected error to be set")
		}

		// Should return command to continue streaming
		if cmd == nil {
			t.Error("Expected command to continue streaming")
		}

		// Verify error is displayed in view
		view := lv.View()
		if !strings.Contains(view, "Error:") {
			t.Error("Expected error to be displayed in view")
		}
	})

	t.Run("no_pod_selected", func(t *testing.T) {
		// Create data provider with no pod
		noPodProvider := &MockKubeoptic{
			selectedPod: nil,
			logBuffer:   []string{},
		}

		lvNoPod := NewLogViewer(noPodProvider, 120, 40)
		view := lvNoPod.View()

		if !strings.Contains(view, "No Pod Selected") {
			t.Error("Expected 'No Pod Selected' message when no pod is selected")
		}
	})

	t.Run("stream_eof_handling", func(t *testing.T) {
		eofMsg := tui.LogChunkMsg{EOF: true}
		_, cmd := lv.Update(eofMsg)

		// Should not return command when EOF is reached
		if cmd != nil {
			t.Error("Expected no command when EOF is reached")
		}
	})
}

func TestLogViewerIntegration_PerformanceWithLargeLogs(t *testing.T) {
	dataProvider := newStreamingMockKubeoptic()
	lv := NewLogViewer(dataProvider, 120, 40)

	t.Run("large_log_volume", func(t *testing.T) {
		// Generate a large number of log entries
		startTime := time.Now()

		for i := 0; i < 1000; i++ {
			logMsg := tui.LogChunkMsg{
				Data: "Log entry " + string(rune(i)) + " with some content to test performance",
			}
			lv.Update(logMsg)
		}

		duration := time.Since(startTime)

		// Should handle 1000 log entries quickly (under 100ms)
		if duration > 100*time.Millisecond {
			t.Errorf("Performance test failed: took %v to process 1000 log entries", duration)
		}

		// Verify buffer management (should not exceed maxLogLines)
		if len(lv.logLines) > maxLogLines {
			t.Errorf("Log buffer exceeded limit: %d > %d", len(lv.logLines), maxLogLines)
		}
	})

	t.Run("search_performance", func(t *testing.T) {
		// Add many log entries
		for i := 0; i < 5000; i++ {
			logData := "INFO Log entry " + string(rune(i%100)) + " testing search performance"
			if i%100 == 0 {
				logData = "ERROR Special log entry " + string(rune(i))
			}
			lv.appendLogData(logData)
		}

		// Test search performance
		startTime := time.Now()
		lv.searchQuery = "ERROR"
		lv.updateSearchResults()
		duration := time.Since(startTime)

		// Search should be fast even with many logs
		if duration > 50*time.Millisecond {
			t.Errorf("Search performance test failed: took %v to search through logs", duration)
		}

		// Should find the ERROR entries
		if len(lv.searchResults) == 0 {
			t.Error("Expected to find ERROR entries in search")
		}
	})
}

func TestLogViewerIntegration_UserInteractionWorkflows(t *testing.T) {
	dataProvider := newStreamingMockKubeoptic()
	lv := NewLogViewer(dataProvider, 120, 40)
	lv.Focus()

	// Add some test data
	lv.appendLogData("INFO Application started")
	lv.appendLogData("ERROR Database connection failed")
	lv.appendLogData("WARN Retrying connection")
	lv.appendLogData("INFO Connected successfully")

	t.Run("complete_search_workflow", func(t *testing.T) {
		// User presses '/' to start search
		searchKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
		lv.Update(searchKey)

		if !lv.searchMode {
			t.Error("Expected search mode to be active after pressing '/'")
		}

		// User types search term
		lv.searchInput.SetValue("ERROR")
		lv.searchQuery = lv.searchInput.Value()
		lv.updateSearchResults()

		// User presses Enter to execute search
		enterKey := tea.KeyMsg{Type: tea.KeyEnter}
		lv.handleSearchMode(enterKey)

		if lv.searchMode {
			t.Error("Expected to exit search mode after pressing Enter")
		}

		if len(lv.searchResults) == 0 {
			t.Error("Expected search results for 'ERROR'")
		}

		// User navigates through results with 'n'
		nextKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
		lv.Update(nextKey)

		// User clears search with Esc
		escKey := tea.KeyMsg{Type: tea.KeyEsc}
		lv.Update(escKey)

		if lv.searchQuery != "" {
			t.Error("Expected search to be cleared after pressing Esc")
		}
	})

	t.Run("navigation_workflow", func(t *testing.T) {
		// User presses 'f' to toggle follow mode
		followKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
		initialFollow := lv.followMode
		_, cmd := lv.Update(followKey)

		// The 'f' key returns a ToggleFollowMsg command, simulate executing it
		if cmd != nil {
			// Execute the returned command to get the ToggleFollowMsg
			msg := cmd()
			if toggleMsg, ok := msg.(tui.ToggleFollowMsg); ok {
				lv.Update(toggleMsg)
			}
		}

		if lv.followMode == initialFollow {
			t.Error("Expected follow mode to toggle")
		}

		// User presses 'w' to toggle line wrapping
		wrapKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")}
		initialWrap := lv.wrapLines
		lv.Update(wrapKey)

		if lv.wrapLines == initialWrap {
			t.Error("Expected wrap mode to toggle")
		}

		// User presses 't' to toggle timestamps
		timeKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
		initialTime := lv.showTimestamps
		lv.Update(timeKey)

		if lv.showTimestamps == initialTime {
			t.Error("Expected timestamp mode to toggle")
		}
	})

	t.Run("window_resize_workflow", func(t *testing.T) {
		// Simulate window resize
		resizeMsg := tea.WindowSizeMsg{Width: 150, Height: 50}
		lv.Update(resizeMsg)

		if lv.width != 150 || lv.height != 50 {
			t.Errorf("Expected size 150x50, got %dx%d", lv.width, lv.height)
		}

		// Verify viewport was resized
		expectedWidth := 150 - 2 // Account for borders
		expectedHeight := 50 - 4 // Account for header/footer

		if lv.viewport.Width != expectedWidth || lv.viewport.Height != expectedHeight {
			t.Errorf("Expected viewport %dx%d, got %dx%d",
				expectedWidth, expectedHeight, lv.viewport.Width, lv.viewport.Height)
		}
	})
}

func TestLogViewerIntegration_RealTimeStreaming(t *testing.T) {
	dataProvider := newStreamingMockKubeoptic()
	lv := NewLogViewer(dataProvider, 120, 40)

	t.Run("continuous_log_streaming", func(t *testing.T) {
		// Simulate continuous log arrival
		logEntries := []string{
			"2023-01-01 10:00:01 INFO Service starting",
			"2023-01-01 10:00:02 INFO Loading modules",
			"2023-01-01 10:00:03 ERROR Module X failed to load",
			"2023-01-01 10:00:04 WARN Retrying module X",
			"2023-01-01 10:00:05 INFO Module X loaded successfully",
			"2023-01-01 10:00:06 INFO All modules loaded",
			"2023-01-01 10:00:07 INFO Service ready",
		}

		for i, entry := range logEntries {
			msg := tui.LogChunkMsg{Data: entry}
			lv.Update(msg)

			// Verify log was added
			if len(lv.logLines) != i+1 {
				t.Errorf("Expected %d log lines after entry %d, got %d",
					i+1, i, len(lv.logLines))
			}

			// In follow mode, should maintain position at bottom
			if lv.followMode {
				// Test that follow mode works (simplified test)
				// In real implementation, we'd verify viewport scroll position
			}
		}

		// Verify all entries are present
		content := lv.renderLogContent()
		for _, entry := range logEntries {
			if !strings.Contains(content, entry) {
				t.Errorf("Expected log content to contain '%s'", entry)
			}
		}
	})

	t.Run("filtered_streaming", func(t *testing.T) {
		// Create a fresh log viewer for this test
		freshProvider := newStreamingMockKubeoptic()
		freshLv := NewLogViewer(freshProvider, 120, 40)

		// Set up a search filter
		freshLv.searchQuery = "ERROR"

		// Stream logs with mix of levels
		logEntries := []string{
			"INFO Normal operation",
			"ERROR Critical failure",
			"INFO Recovery attempt",
			"ERROR Secondary failure",
			"INFO System stable",
		}

		for _, entry := range logEntries {
			msg := tui.LogChunkMsg{Data: entry}
			freshLv.Update(msg)
		}

		// Update search results
		freshLv.updateSearchResults()

		// Should find 2 ERROR entries
		expectedErrors := 2
		if len(freshLv.searchResults) != expectedErrors {
			t.Errorf("Expected %d ERROR entries, got %d",
				expectedErrors, len(freshLv.searchResults))
		}
	})
}
