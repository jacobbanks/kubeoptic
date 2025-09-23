package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Style helper utilities for common patterns in kubeoptic TUI

// JoinHorizontal joins styles horizontally with optional separator
func JoinHorizontal(separator string, styles ...string) string {
	if len(styles) == 0 {
		return ""
	}
	return strings.Join(styles, separator)
}

// JoinVertical joins styles vertically
func JoinVertical(styles ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, styles...)
}

// CenterText centers text within a given width
func CenterText(text string, width int, theme Theme) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Foreground).
		Render(text)
}

// TruncateText truncates text to fit within specified width
func TruncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	return text[:maxWidth-3] + "..."
}

// PadToWidth pads text to a specific width
func PadToWidth(text string, width int) string {
	if len(text) >= width {
		return TruncateText(text, width)
	}
	return text + strings.Repeat(" ", width-len(text))
}

// FormatDuration formats a duration string with appropriate styling
func FormatDuration(duration string, theme Theme) string {
	return lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Render(duration)
}

// FormatBytes formats byte count in human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CreateProgressBar creates a simple progress bar
func CreateProgressBar(current, total int, width int, theme Theme) string {
	if total == 0 {
		return strings.Repeat("─", width)
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("─", width-filled)
	return lipgloss.NewStyle().
		Foreground(theme.Primary).
		Render(bar)
}

// StyleWithFocus applies focus styling conditionally
func StyleWithFocus(baseStyle lipgloss.Style, focused bool, theme Theme) lipgloss.Style {
	if focused {
		return baseStyle.
			BorderForeground(theme.Primary).
			Foreground(theme.Primary)
	}
	return baseStyle.
		BorderForeground(theme.Border).
		Foreground(theme.Foreground)
}

// CreateBadge creates a styled badge for status indicators
func CreateBadge(text string, color lipgloss.Color, theme Theme) string {
	return lipgloss.NewStyle().
		Background(color).
		Foreground(theme.Background).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// CreateSeparator creates a horizontal separator line
func CreateSeparator(width int, theme Theme) string {
	return lipgloss.NewStyle().
		Width(width).
		Foreground(theme.Border).
		Render(strings.Repeat("─", width))
}

// FormatKeyValue formats key-value pairs with consistent styling
func FormatKeyValue(key, value string, keyColor, valueColor lipgloss.Color) string {
	keyStyle := lipgloss.NewStyle().Foreground(keyColor).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(valueColor)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render(key+": "),
		valueStyle.Render(value))
}

// CreateTable creates a simple table layout
func CreateTable(headers []string, rows [][]string, theme Theme, width int) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colCount := len(headers)
	colWidth := (width - (colCount - 1)) / colCount

	// Create header
	headerCells := make([]string, len(headers))
	for i, header := range headers {
		headerCells[i] = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Width(colWidth).
			Align(lipgloss.Center).
			Render(TruncateText(header, colWidth))
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Left, headerCells...)

	// Create separator
	separator := CreateSeparator(width, theme)

	// Create data rows
	var dataRows []string
	for _, row := range rows {
		cells := make([]string, len(headers))
		for i, cell := range row {
			if i >= len(headers) {
				break
			}
			cells[i] = lipgloss.NewStyle().
				Foreground(theme.Foreground).
				Width(colWidth).
				Render(TruncateText(cell, colWidth))
		}
		dataRows = append(dataRows, lipgloss.JoinHorizontal(lipgloss.Left, cells...))
	}

	// Combine all parts
	parts := []string{headerRow, separator}
	parts = append(parts, dataRows...)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ApplyThemeToText applies theme colors to text based on keywords
func ApplyThemeToText(text string, theme Theme) string {
	// Apply syntax highlighting for common log patterns
	text = strings.ReplaceAll(text, "ERROR",
		lipgloss.NewStyle().Foreground(theme.Error).Bold(true).Render("ERROR"))
	text = strings.ReplaceAll(text, "WARN",
		lipgloss.NewStyle().Foreground(theme.Warning).Bold(true).Render("WARN"))
	text = strings.ReplaceAll(text, "INFO",
		lipgloss.NewStyle().Foreground(theme.Info).Render("INFO"))
	text = strings.ReplaceAll(text, "DEBUG",
		lipgloss.NewStyle().Foreground(Gray).Render("DEBUG"))

	return text
}

// WrapText wraps text to fit within specified width
func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed width, start new line
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
