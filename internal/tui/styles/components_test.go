package styles

import (
	"strings"
	"testing"
)

func TestHeaderStyles(t *testing.T) {
	theme := DefaultTheme()
	width := 100

	style := HeaderStyles(theme, width)
	result := style.Render("Test Header")

	if !strings.Contains(result, "Test Header") {
		t.Error("HeaderStyles should render the header text")
	}
}

func TestStatusBarStyles(t *testing.T) {
	theme := DefaultTheme()
	width := 100

	style := StatusBarStyles(theme, width)
	result := style.Render("Status: Connected")

	if !strings.Contains(result, "Status: Connected") {
		t.Error("StatusBarStyles should render the status text")
	}
}

func TestNewListStyles(t *testing.T) {
	theme := DefaultTheme()
	width := 50
	height := 20

	// Test focused list styles
	focusedStyles := NewListStyles(theme, width, height, true)
	if focusedStyles.Container.Render("") == "" {
		t.Error("Focused list container style should be valid")
	}

	// Test unfocused list styles
	unfocusedStyles := NewListStyles(theme, width, height, false)
	if unfocusedStyles.Container.Render("") == "" {
		t.Error("Unfocused list container style should be valid")
	}

	// Test all style components
	testText := "test item"

	if !strings.Contains(focusedStyles.Title.Render(testText), testText) {
		t.Error("List title style should render text")
	}
	if !strings.Contains(focusedStyles.Item.Render(testText), testText) {
		t.Error("List item style should render text")
	}
	if !strings.Contains(focusedStyles.SelectedItem.Render(testText), testText) {
		t.Error("List selected item style should render text")
	}
	if !strings.Contains(focusedStyles.FilteredItem.Render(testText), testText) {
		t.Error("List filtered item style should render text")
	}
	if !strings.Contains(focusedStyles.EmptyState.Render(testText), testText) {
		t.Error("List empty state style should render text")
	}
}

func TestNewLogViewerStyles(t *testing.T) {
	theme := DefaultTheme()
	width := 80
	height := 30

	// Test focused log viewer styles
	focusedStyles := NewLogViewerStyles(theme, width, height, true)
	if focusedStyles.Container.Render("") == "" {
		t.Error("Focused log viewer container style should be valid")
	}

	// Test unfocused log viewer styles
	unfocusedStyles := NewLogViewerStyles(theme, width, height, false)
	if unfocusedStyles.Container.Render("") == "" {
		t.Error("Unfocused log viewer container style should be valid")
	}

	// Test all log style components
	testText := "test log"

	if !strings.Contains(focusedStyles.Title.Render(testText), testText) {
		t.Error("Log viewer title style should render text")
	}
	if !strings.Contains(focusedStyles.LogLine.Render(testText), testText) {
		t.Error("Log line style should render text")
	}
	if !strings.Contains(focusedStyles.ErrorLog.Render(testText), testText) {
		t.Error("Error log style should render text")
	}
	if !strings.Contains(focusedStyles.WarningLog.Render(testText), testText) {
		t.Error("Warning log style should render text")
	}
	if !strings.Contains(focusedStyles.InfoLog.Render(testText), testText) {
		t.Error("Info log style should render text")
	}
	if !strings.Contains(focusedStyles.DebugLog.Render(testText), testText) {
		t.Error("Debug log style should render text")
	}
	if !strings.Contains(focusedStyles.Timestamp.Render(testText), testText) {
		t.Error("Timestamp style should render text")
	}
}

func TestPodStatusStyles(t *testing.T) {
	theme := DefaultTheme()
	statuses := []string{"Running", "Pending", "Failed", "Succeeded", "Terminated", "Unknown"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			style := PodStatusStyles(status, theme)
			result := style.Render(status)

			if !strings.Contains(result, status) {
				t.Errorf("PodStatusStyles should render status %q", status)
			}
		})
	}
}

func TestContextStyles(t *testing.T) {
	theme := DefaultTheme()
	contextName := "production"

	// Test active context
	activeStyle := ContextStyles(theme, true)
	activeResult := activeStyle.Render(contextName)
	if !strings.Contains(activeResult, contextName) {
		t.Error("Active context style should render context name")
	}

	// Test inactive context
	inactiveStyle := ContextStyles(theme, false)
	inactiveResult := inactiveStyle.Render(contextName)
	if !strings.Contains(inactiveResult, contextName) {
		t.Error("Inactive context style should render context name")
	}
}

func TestNamespaceStyles(t *testing.T) {
	theme := DefaultTheme()
	namespaceName := "kube-system"

	// Test selected namespace
	selectedStyle := NamespaceStyles(namespaceName, theme, true)
	selectedResult := selectedStyle.Render(namespaceName)
	if !strings.Contains(selectedResult, namespaceName) {
		t.Error("Selected namespace style should render namespace name")
	}

	// Test unselected namespace
	unselectedStyle := NamespaceStyles(namespaceName, theme, false)
	unselectedResult := unselectedStyle.Render(namespaceName)
	if !strings.Contains(unselectedResult, namespaceName) {
		t.Error("Unselected namespace style should render namespace name")
	}
}

func TestSearchInputStyles(t *testing.T) {
	theme := DefaultTheme()
	searchText := "search term"

	// Test focused search input
	focusedStyle := SearchInputStyles(theme, true)
	focusedResult := focusedStyle.Render(searchText)
	if !strings.Contains(focusedResult, searchText) {
		t.Error("Focused search input style should render search text")
	}

	// Test unfocused search input
	unfocusedStyle := SearchInputStyles(theme, false)
	unfocusedResult := unfocusedStyle.Render(searchText)
	if !strings.Contains(unfocusedResult, searchText) {
		t.Error("Unfocused search input style should render search text")
	}
}

func TestHelpStyles(t *testing.T) {
	theme := DefaultTheme()
	helpText := "Press 'q' to quit"

	style := HelpStyles(theme)
	result := style.Render(helpText)

	if !strings.Contains(result, helpText) {
		t.Error("Help style should render help text")
	}
}

func TestErrorMessageStyles(t *testing.T) {
	theme := DefaultTheme()
	width := 100
	errorText := "Connection failed"

	style := ErrorMessageStyles(theme, width)
	result := style.Render(errorText)

	if !strings.Contains(result, errorText) {
		t.Error("Error message style should render error text")
	}
}

func TestStyleConsistency(t *testing.T) {
	theme := DefaultTheme()

	// Test that multiple calls to the same style function return consistent results
	width := 100

	style1 := HeaderStyles(theme, width)
	style2 := HeaderStyles(theme, width)

	testText := "Consistency Test"
	result1 := style1.Render(testText)
	result2 := style2.Render(testText)

	if result1 != result2 {
		t.Error("HeaderStyles should return consistent results for same parameters")
	}
}

func TestStylesWithDifferentThemes(t *testing.T) {
	defaultTheme := DefaultTheme()
	darkTheme := DarkTheme()

	testText := "Theme Test"

	// Test that different themes produce different styled output
	defaultHeader := HeaderStyles(defaultTheme, 100).Render(testText)
	darkHeader := HeaderStyles(darkTheme, 100).Render(testText)

	// Both should contain the test text
	if !strings.Contains(defaultHeader, testText) {
		t.Error("Default theme header should contain test text")
	}
	if !strings.Contains(darkHeader, testText) {
		t.Error("Dark theme header should contain test text")
	}
}

func TestListStylesWithDifferentFocusStates(t *testing.T) {
	theme := DefaultTheme()
	width, height := 50, 20

	focusedStyles := NewListStyles(theme, width, height, true)
	unfocusedStyles := NewListStyles(theme, width, height, false)

	testText := "Focus Test"

	// Test that focused and unfocused containers render differently
	focusedContainer := focusedStyles.Container.Render(testText)
	unfocusedContainer := unfocusedStyles.Container.Render(testText)

	if !strings.Contains(focusedContainer, testText) {
		t.Error("Focused container should contain test text")
	}
	if !strings.Contains(unfocusedContainer, testText) {
		t.Error("Unfocused container should contain test text")
	}
}

func TestLogViewerStylesWithDifferentFocusStates(t *testing.T) {
	theme := DefaultTheme()
	width, height := 80, 30

	focusedStyles := NewLogViewerStyles(theme, width, height, true)
	unfocusedStyles := NewLogViewerStyles(theme, width, height, false)

	testText := "Log Focus Test"

	// Test that focused and unfocused containers render
	focusedContainer := focusedStyles.Container.Render(testText)
	unfocusedContainer := unfocusedStyles.Container.Render(testText)

	if !strings.Contains(focusedContainer, testText) {
		t.Error("Focused log container should contain test text")
	}
	if !strings.Contains(unfocusedContainer, testText) {
		t.Error("Unfocused log container should contain test text")
	}
}

func BenchmarkHeaderStyles(b *testing.B) {
	theme := DefaultTheme()
	width := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HeaderStyles(theme, width)
	}
}

func BenchmarkNewListStyles(b *testing.B) {
	theme := DefaultTheme()
	width, height := 50, 20

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewListStyles(theme, width, height, true)
	}
}

func BenchmarkPodStatusStyles(b *testing.B) {
	theme := DefaultTheme()
	status := "Running"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PodStatusStyles(status, theme)
	}
}
