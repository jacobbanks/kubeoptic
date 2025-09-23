package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestJoinHorizontal(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		styles    []string
		expected  string
	}{
		{"empty styles", "|", []string{}, ""},
		{"single style", "|", []string{"hello"}, "hello"},
		{"multiple styles", "|", []string{"hello", "world"}, "hello|world"},
		{"with spaces", " - ", []string{"one", "two", "three"}, "one - two - three"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := JoinHorizontal(test.separator, test.styles...)
			if result != test.expected {
				t.Errorf("JoinHorizontal() = %q, want %q", result, test.expected)
			}
		})
	}
}

func TestJoinVertical(t *testing.T) {
	styles := []string{"line1", "line2", "line3"}
	result := JoinVertical(styles...)

	if !strings.Contains(result, "line1") {
		t.Error("Result should contain line1")
	}
	if !strings.Contains(result, "line2") {
		t.Error("Result should contain line2")
	}
	if !strings.Contains(result, "line3") {
		t.Error("Result should contain line3")
	}
}

func TestCenterText(t *testing.T) {
	theme := DefaultTheme()
	text := "hello"
	width := 20

	result := CenterText(text, width, theme)

	// The result should contain the original text
	if !strings.Contains(result, text) {
		t.Errorf("CenterText should contain original text %q", text)
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact width", "hello", 5, "hello"},
		{"truncation needed", "hello world", 8, "hello..."},
		{"very short width", "hello", 3, "..."},
		{"zero width", "hello", 0, ""},
		{"width 1", "hello", 1, "."},
		{"width 2", "hello", 2, ".."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := TruncateText(test.text, test.maxWidth)
			if result != test.expected {
				t.Errorf("TruncateText(%q, %d) = %q, want %q", test.text, test.maxWidth, result, test.expected)
			}
			if len(result) > test.maxWidth {
				t.Errorf("TruncateText result length %d exceeds maxWidth %d", len(result), test.maxWidth)
			}
		})
	}
}

func TestPadToWidth(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{"pad needed", "hello", 10, "hello     "},
		{"no pad needed", "hello world", 5, "he..."},
		{"exact width", "hello", 5, "hello"},
		{"truncation", "hello world", 8, "hello..."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := PadToWidth(test.text, test.width)
			if len(result) != test.width {
				t.Errorf("PadToWidth result length %d, want %d", len(result), test.width)
			}
			if result != test.expected {
				t.Errorf("PadToWidth(%q, %d) = %q, want %q", test.text, test.width, result, test.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	theme := DefaultTheme()
	duration := "5m30s"

	result := FormatDuration(duration, theme)

	if !strings.Contains(result, duration) {
		t.Errorf("FormatDuration should contain original duration %q", duration)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := FormatBytes(test.bytes)
			if result != test.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", test.bytes, result, test.expected)
			}
		})
	}
}

func TestCreateProgressBar(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		name    string
		current int
		total   int
		width   int
	}{
		{"empty progress", 0, 100, 10},
		{"half progress", 50, 100, 10},
		{"full progress", 100, 100, 10},
		{"over progress", 150, 100, 10},
		{"zero total", 50, 0, 10},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CreateProgressBar(test.current, test.total, test.width, theme)

			// Check that result has correct length (note: may have ANSI codes)
			if result == "" {
				t.Error("CreateProgressBar should not return empty string")
			}
		})
	}
}

func TestStyleWithFocus(t *testing.T) {
	theme := DefaultTheme()
	baseStyle := lipgloss.NewStyle().Foreground(theme.Foreground)

	focusedStyle := StyleWithFocus(baseStyle, true, theme)
	unfocusedStyle := StyleWithFocus(baseStyle, false, theme)

	// Test that both styles are valid (render without error)
	testText := "test"
	focusedResult := focusedStyle.Render(testText)
	unfocusedResult := unfocusedStyle.Render(testText)

	if focusedResult == "" {
		t.Error("Focused style should render text")
	}
	if unfocusedResult == "" {
		t.Error("Unfocused style should render text")
	}

	// Both should contain the original text
	if !strings.Contains(focusedResult, testText) {
		t.Error("Focused style should contain original text")
	}
	if !strings.Contains(unfocusedResult, testText) {
		t.Error("Unfocused style should contain original text")
	}
}

func TestCreateBadge(t *testing.T) {
	theme := DefaultTheme()
	text := "RUNNING"
	color := StatusRunning

	result := CreateBadge(text, color, theme)

	if !strings.Contains(result, text) {
		t.Errorf("CreateBadge should contain text %q", text)
	}
}

func TestCreateSeparator(t *testing.T) {
	theme := DefaultTheme()
	width := 20

	result := CreateSeparator(width, theme)

	if result == "" {
		t.Error("CreateSeparator should not return empty string")
	}
}

func TestFormatKeyValue(t *testing.T) {
	key := "Status"
	value := "Running"
	keyColor := PrimaryBlue
	valueColor := StatusRunning

	result := FormatKeyValue(key, value, keyColor, valueColor)

	if !strings.Contains(result, key) {
		t.Errorf("FormatKeyValue should contain key %q", key)
	}
	if !strings.Contains(result, value) {
		t.Errorf("FormatKeyValue should contain value %q", value)
	}
}

func TestCreateTable(t *testing.T) {
	theme := DefaultTheme()
	headers := []string{"Name", "Status", "Age"}
	rows := [][]string{
		{"pod-1", "Running", "5m"},
		{"pod-2", "Pending", "1m"},
	}
	width := 50

	result := CreateTable(headers, rows, theme, width)

	// Check that headers are included
	for _, header := range headers {
		if !strings.Contains(result, header) {
			t.Errorf("CreateTable should contain header %q", header)
		}
	}

	// Check that row data is included
	if !strings.Contains(result, "pod-1") {
		t.Error("CreateTable should contain row data")
	}
}

func TestCreateTableEmptyInput(t *testing.T) {
	theme := DefaultTheme()

	// Test empty headers
	result1 := CreateTable([]string{}, [][]string{{"data"}}, theme, 50)
	if result1 != "" {
		t.Error("CreateTable with empty headers should return empty string")
	}

	// Test empty rows
	result2 := CreateTable([]string{"header"}, [][]string{}, theme, 50)
	if result2 != "" {
		t.Error("CreateTable with empty rows should return empty string")
	}
}

func TestApplyThemeToText(t *testing.T) {
	theme := DefaultTheme()
	text := "This is an ERROR message with WARN and INFO levels"

	result := ApplyThemeToText(text, theme)

	// The result should still contain the original keywords
	if !strings.Contains(result, "ERROR") {
		t.Error("ApplyThemeToText should preserve ERROR keyword")
	}
	if !strings.Contains(result, "WARN") {
		t.Error("ApplyThemeToText should preserve WARN keyword")
	}
	if !strings.Contains(result, "INFO") {
		t.Error("ApplyThemeToText should preserve INFO keyword")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected number of lines
	}{
		{"short text", "hello", 10, 1},
		{"exact width", "hello world", 11, 1},
		{"needs wrapping", "hello world this is a test", 10, 3},
		{"zero width", "hello", 0, 1},
		{"empty text", "", 10, 1},
		{"single word longer than width", "verylongword", 5, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := WrapText(test.text, test.width)

			if len(result) != test.expected {
				t.Errorf("WrapText(%q, %d) returned %d lines, want %d", test.text, test.width, len(result), test.expected)
			}

			// Test that all original words are preserved
			originalWords := strings.Fields(test.text)
			var resultWords []string
			for _, line := range result {
				resultWords = append(resultWords, strings.Fields(line)...)
			}

			if len(originalWords) != len(resultWords) {
				t.Errorf("WrapText should preserve all words: original %d, result %d", len(originalWords), len(resultWords))
			}
		})
	}
}

func BenchmarkTruncateText(b *testing.B) {
	text := "This is a long text that needs to be truncated"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TruncateText(text, 20)
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	bytes := int64(1024 * 1024 * 1024) // 1GB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatBytes(bytes)
	}
}

func BenchmarkWrapText(b *testing.B) {
	text := "This is a long text that needs to be wrapped into multiple lines"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapText(text, 20)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
