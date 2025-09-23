package messages

import (
	"kubeoptic/internal/services"
)

// TUI Message Types for Bubble Tea Update() pattern

// Navigation Messages
type NavigateMsg struct {
	Direction NavigationDirection
}

type NavigationDirection int

const (
	NavigateUp NavigationDirection = iota
	NavigateDown
	NavigateLeft
	NavigateRight
	NavigateBack
	NavigateForward
)

// Data Messages
type ContextsLoadedMsg struct {
	Contexts []services.Context
	Error    error
}

type NamespacesLoadedMsg struct {
	Namespaces []string
	Error      error
}

type PodsLoadedMsg struct {
	Pods  []services.Pod
	Error error
}

type ContextSelectedMsg struct {
	Context *services.Context
}

type PodSelectedMsg struct {
	Pod *services.Pod
}

// Log Messages
type LogChunkMsg struct {
	Data  string
	EOF   bool
	Error error
}

type LogStreamStartedMsg struct {
	Pod *services.Pod
}

type LogStreamStoppedMsg struct {
	Reason string
}

type ToggleFollowMsg struct{}

// Search Messages
type SearchQueryChangedMsg struct {
	Query string
}

type SearchResultsMsg struct {
	Results []services.Pod
	Query   string
}

type ClearSearchMsg struct{}

// UI Messages
type WindowResizeMsg struct {
	Width  int
	Height int
}

type FocusChangedMsg struct {
	Component string
	Focused   bool
}

type ErrorMsg struct {
	Error   error
	Context string
}

type StatusMsg struct {
	Message string
	Type    StatusType
}

type ClearStatusMsg struct{}

// System Messages
type InitCompleteMsg struct{}

type ShutdownMsg struct{}

type RefreshDataMsg struct{}

// Loading state messages
type LoadingStartedMsg struct {
	Component string
}

type LoadingCompletedMsg struct {
	Component string
}

// StatusType represents the type of status message
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// String returns the string representation of StatusType
func (s StatusType) String() string {
	switch s {
	case StatusInfo:
		return "info"
	case StatusSuccess:
		return "success"
	case StatusWarning:
		return "warning"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}
