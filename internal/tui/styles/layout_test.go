package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewLayoutConfig(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small screen", 60, 20},
		{"medium screen", 100, 30},
		{"large screen", 200, 50},
		{"minimum size", 10, 5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := NewLayoutConfig(test.width, test.height)

			if config.Width != test.width {
				t.Errorf("Width = %d, want %d", config.Width, test.width)
			}
			if config.Height != test.height {
				t.Errorf("Height = %d, want %d", config.Height, test.height)
			}

			// Test that screen size categories are mutually exclusive
			categoryCount := 0
			if config.IsSmallScreen {
				categoryCount++
			}
			if config.IsMediumScreen {
				categoryCount++
			}
			if config.IsLargeScreen {
				categoryCount++
			}

			if categoryCount != 1 {
				t.Errorf("Expected exactly one screen size category to be true, got %d", categoryCount)
			}
		})
	}
}

func TestLayoutConfigScreenSizes(t *testing.T) {
	tests := []struct {
		width          int
		expectedSmall  bool
		expectedMedium bool
		expectedLarge  bool
	}{
		{60, true, false, false},
		{80, true, false, false},
		{100, false, true, false},
		{120, false, true, false},
		{140, false, false, true},
		{200, false, false, true},
	}

	for _, test := range tests {
		t.Run("width_"+string(rune(test.width)), func(t *testing.T) {
			config := NewLayoutConfig(test.width, 30)

			if config.IsSmallScreen != test.expectedSmall {
				t.Errorf("IsSmallScreen = %v, want %v", config.IsSmallScreen, test.expectedSmall)
			}
			if config.IsMediumScreen != test.expectedMedium {
				t.Errorf("IsMediumScreen = %v, want %v", config.IsMediumScreen, test.expectedMedium)
			}
			if config.IsLargeScreen != test.expectedLarge {
				t.Errorf("IsLargeScreen = %v, want %v", config.IsLargeScreen, test.expectedLarge)
			}
		})
	}
}

func TestLayoutConfigPanelWidths(t *testing.T) {
	// Test small screen
	smallConfig := NewLayoutConfig(70, 25)
	if smallConfig.ContextListWidth <= 0 {
		t.Error("ContextListWidth should be positive")
	}
	if smallConfig.NamespaceListWidth <= 0 {
		t.Error("NamespaceListWidth should be positive")
	}
	if smallConfig.PodListWidth <= 0 {
		t.Error("PodListWidth should be positive")
	}

	// Test medium screen
	mediumConfig := NewLayoutConfig(100, 25)
	totalWidth := mediumConfig.ContextListWidth + mediumConfig.NamespaceListWidth + mediumConfig.PodListWidth
	if totalWidth > mediumConfig.Width {
		t.Errorf("Total panel width %d exceeds screen width %d", totalWidth, mediumConfig.Width)
	}

	// Test large screen
	largeConfig := NewLayoutConfig(180, 40)
	totalWidthLarge := largeConfig.ContextListWidth + largeConfig.NamespaceListWidth + largeConfig.PodListWidth
	if totalWidthLarge > largeConfig.Width {
		t.Errorf("Total panel width %d exceeds screen width %d", totalWidthLarge, largeConfig.Width)
	}
}

func TestGetPanelWidth(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{60, 56},  // Small screen: width - 4
		{90, 30},  // Medium screen: width / 3
		{120, 40}, // Medium screen: width / 3
		{180, 60}, // Large screen: width / 3
	}

	for _, test := range tests {
		t.Run("width_"+string(rune(test.width)), func(t *testing.T) {
			config := NewLayoutConfig(test.width, 30)
			result := config.GetPanelWidth()
			if result != test.expected {
				t.Errorf("GetPanelWidth() = %d, want %d", result, test.expected)
			}
		})
	}
}

func TestGetMainViewHeight(t *testing.T) {
	tests := []struct {
		height   int
		expected int
	}{
		{20, 16}, // 20 - HeaderHeight(3) - StatusBarHeight(1)
		{30, 26}, // 30 - HeaderHeight(3) - StatusBarHeight(1)
		{50, 46}, // 50 - HeaderHeight(3) - StatusBarHeight(1)
	}

	for _, test := range tests {
		t.Run("height_"+string(rune(test.height)), func(t *testing.T) {
			config := NewLayoutConfig(100, test.height)
			result := config.GetMainViewHeight()
			if result != test.expected {
				t.Errorf("GetMainViewHeight() = %d, want %d", result, test.expected)
			}
		})
	}
}

func TestGetListHeight(t *testing.T) {
	config := NewLayoutConfig(100, 30)
	height := config.GetListHeight()

	if height < MinListHeight {
		t.Errorf("GetListHeight() = %d, should be at least MinListHeight(%d)", height, MinListHeight)
	}

	mainHeight := config.GetMainViewHeight()
	if height > mainHeight {
		t.Errorf("GetListHeight() = %d, should not exceed main view height(%d)", height, mainHeight)
	}
}

func TestShouldStackVertically(t *testing.T) {
	smallConfig := NewLayoutConfig(60, 25)
	if !smallConfig.ShouldStackVertically() {
		t.Error("Small screen should stack vertically")
	}

	largeConfig := NewLayoutConfig(150, 40)
	if largeConfig.ShouldStackVertically() {
		t.Error("Large screen should not stack vertically")
	}
}

func TestCalculateThreePanelLayout(t *testing.T) {
	tests := []struct {
		width int
		name  string
	}{
		{60, "small"},
		{100, "medium"},
		{180, "large"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := NewLayoutConfig(test.width, 30)
			left, middle, right := config.CalculateThreePanelLayout()

			if left <= 0 || middle <= 0 || right <= 0 {
				t.Errorf("All panel widths should be positive: left=%d, middle=%d, right=%d", left, middle, right)
			}

			if config.IsSmallScreen {
				// On small screens, all panels should have the same width (full width for stacking)
				if left != middle || middle != right {
					t.Errorf("Small screen panels should have equal widths: left=%d, middle=%d, right=%d", left, middle, right)
				}
			} else {
				// On larger screens, total width should not exceed available space
				totalWidth := left + middle + right + 6 // Account for borders/spacing
				if totalWidth > test.width {
					t.Errorf("Total panel width %d exceeds screen width %d", totalWidth, test.width)
				}
			}
		})
	}
}

func TestGetBorderStyle(t *testing.T) {
	smallConfig := NewLayoutConfig(60, 25)
	smallBorder := smallConfig.GetBorderStyle()

	largeConfig := NewLayoutConfig(150, 40)
	largeBorder := largeConfig.GetBorderStyle()

	// Test that we get valid border objects
	if smallBorder == (lipgloss.Border{}) {
		t.Error("Small screen should return a valid border")
	}
	if largeBorder == (lipgloss.Border{}) {
		t.Error("Large screen should return a valid border")
	}
}

func TestGetBorders(t *testing.T) {
	borders := GetBorders()

	// Test that all border types are returned
	if borders.None == (lipgloss.Border{}) {
		t.Error("None border should be valid")
	}
	if borders.Round == (lipgloss.Border{}) {
		t.Error("Round border should be valid")
	}
	if borders.Double == (lipgloss.Border{}) {
		t.Error("Double border should be valid")
	}
}

func TestLayoutConstants(t *testing.T) {
	// Test that constants have reasonable values
	if PaddingSmall >= PaddingMedium {
		t.Error("PaddingSmall should be less than PaddingMedium")
	}
	if PaddingMedium >= PaddingLarge {
		t.Error("PaddingMedium should be less than PaddingLarge")
	}

	if MarginSmall >= MarginMedium {
		t.Error("MarginSmall should be less than MarginMedium")
	}
	if MarginMedium >= MarginLarge {
		t.Error("MarginMedium should be less than MarginLarge")
	}

	if MinListHeight <= 0 {
		t.Error("MinListHeight should be positive")
	}
	if MinListWidth <= 0 {
		t.Error("MinListWidth should be positive")
	}

	if SmallScreenWidth >= MediumScreenWidth {
		t.Error("SmallScreenWidth should be less than MediumScreenWidth")
	}
	if MediumScreenWidth >= LargeScreenWidth {
		t.Error("MediumScreenWidth should be less than LargeScreenWidth")
	}
}

func BenchmarkNewLayoutConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewLayoutConfig(120, 30)
	}
}

func BenchmarkCalculateThreePanelLayout(b *testing.B) {
	config := NewLayoutConfig(120, 30)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.CalculateThreePanelLayout()
	}
}
