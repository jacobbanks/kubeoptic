package tui

import (
	"testing"
)

func TestStatusType_String(t *testing.T) {
	tests := []struct {
		status   StatusType
		expected string
	}{
		{StatusInfo, "info"},
		{StatusSuccess, "success"},
		{StatusWarning, "warning"},
		{StatusError, "error"},
		{StatusType(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.status.String()
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestNavigationDirection_Values(t *testing.T) {
	// Test that navigation direction constants are defined
	directions := []NavigationDirection{
		NavigateUp,
		NavigateDown,
		NavigateLeft,
		NavigateRight,
		NavigateBack,
		NavigateForward,
	}

	// Ensure each direction has a unique value
	seen := make(map[NavigationDirection]bool)
	for _, dir := range directions {
		if seen[dir] {
			t.Errorf("Duplicate navigation direction value: %d", dir)
		}
		seen[dir] = true
	}
}

func TestMessageTypes_Creation(t *testing.T) {
	// Test that message types can be created without panics

	t.Run("NavigateMsg", func(t *testing.T) {
		msg := NavigateMsg{Direction: NavigateUp}
		if msg.Direction != NavigateUp {
			t.Error("NavigateMsg creation failed")
		}
	})

	t.Run("ErrorMsg", func(t *testing.T) {
		msg := ErrorMsg{Error: nil, Context: "test"}
		if msg.Context != "test" {
			t.Error("ErrorMsg creation failed")
		}
	})

	t.Run("StatusMsg", func(t *testing.T) {
		msg := StatusMsg{Message: "test", Type: StatusInfo}
		if msg.Message != "test" || msg.Type != StatusInfo {
			t.Error("StatusMsg creation failed")
		}
	})

	t.Run("LogChunkMsg", func(t *testing.T) {
		msg := LogChunkMsg{Data: "test log", EOF: false, Error: nil}
		if msg.Data != "test log" || msg.EOF != false {
			t.Error("LogChunkMsg creation failed")
		}
	})

	t.Run("SearchQueryChangedMsg", func(t *testing.T) {
		msg := SearchQueryChangedMsg{Query: "test query"}
		if msg.Query != "test query" {
			t.Error("SearchQueryChangedMsg creation failed")
		}
	})

	t.Run("WindowResizeMsg", func(t *testing.T) {
		msg := WindowResizeMsg{Width: 80, Height: 24}
		if msg.Width != 80 || msg.Height != 24 {
			t.Error("WindowResizeMsg creation failed")
		}
	})

	t.Run("FocusChangedMsg", func(t *testing.T) {
		msg := FocusChangedMsg{Component: "podlist", Focused: true}
		if msg.Component != "podlist" || msg.Focused != true {
			t.Error("FocusChangedMsg creation failed")
		}
	})
}

func TestStatusType_Coverage(t *testing.T) {
	// Ensure all status types are covered in String() method
	allStatuses := []StatusType{StatusInfo, StatusSuccess, StatusWarning, StatusError}

	for _, status := range allStatuses {
		result := status.String()
		if result == "unknown" {
			t.Errorf("Status type %d should not return 'unknown'", status)
		}
		if result == "" {
			t.Errorf("Status type %d should not return empty string", status)
		}
	}
}
