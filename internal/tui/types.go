package tui

// Re-export message types from messages package to maintain compatibility
// This allows existing code to continue using tui.MessageType while
// avoiding import cycles

import (
	"kubeoptic/internal/messages"
)

// StatusType represents the type of status message
type StatusType = messages.StatusType

const (
	StatusInfo    = messages.StatusInfo
	StatusSuccess = messages.StatusSuccess
	StatusWarning = messages.StatusWarning
	StatusError   = messages.StatusError
)

// Message Types - re-exported from messages package
type NavigateMsg = messages.NavigateMsg
type NavigationDirection = messages.NavigationDirection

const (
	NavigateUp      = messages.NavigateUp
	NavigateDown    = messages.NavigateDown
	NavigateLeft    = messages.NavigateLeft
	NavigateRight   = messages.NavigateRight
	NavigateBack    = messages.NavigateBack
	NavigateForward = messages.NavigateForward
)

// Data Messages
type ContextsLoadedMsg = messages.ContextsLoadedMsg
type NamespacesLoadedMsg = messages.NamespacesLoadedMsg
type PodsLoadedMsg = messages.PodsLoadedMsg
type ContextSelectedMsg = messages.ContextSelectedMsg
type PodSelectedMsg = messages.PodSelectedMsg

// Log Messages
type LogChunkMsg = messages.LogChunkMsg
type LogStreamStartedMsg = messages.LogStreamStartedMsg
type LogStreamStoppedMsg = messages.LogStreamStoppedMsg
type ToggleFollowMsg = messages.ToggleFollowMsg

// Additional log operation messages
type SaveLogsMsg = messages.SaveLogsMsg
type ToggleWrapMsg = messages.ToggleWrapMsg

// Search Messages
type SearchQueryChangedMsg = messages.SearchQueryChangedMsg
type SearchResultsMsg = messages.SearchResultsMsg
type ClearSearchMsg = messages.ClearSearchMsg

// UI Messages
type WindowResizeMsg = messages.WindowResizeMsg
type FocusChangedMsg = messages.FocusChangedMsg
type ErrorMsg = messages.ErrorMsg
type StatusMsg = messages.StatusMsg
type ClearStatusMsg = messages.ClearStatusMsg

// System Messages
type InitCompleteMsg = messages.InitCompleteMsg
type ShutdownMsg = messages.ShutdownMsg
type RefreshDataMsg = messages.RefreshDataMsg
type LoadingStartedMsg = messages.LoadingStartedMsg
type LoadingCompletedMsg = messages.LoadingCompletedMsg
