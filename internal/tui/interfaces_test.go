package tui

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
)

// Test that all interfaces are properly defined
func TestInterfaces_Definition(t *testing.T) {
	t.Run("ComponentRenderer", func(t *testing.T) {
		var _ ComponentRenderer = (*mockComponentRenderer)(nil)
	})

	t.Run("AppStateManager", func(t *testing.T) {
		var _ AppStateManager = (*mockAppStateManager)(nil)
	})

	t.Run("EventHandler", func(t *testing.T) {
		var _ EventHandler = (*mockEventHandler)(nil)
	})

	t.Run("DataProvider", func(t *testing.T) {
		var _ DataProvider = (*mockDataProvider)(nil)
	})

	t.Run("Focusable", func(t *testing.T) {
		var _ Focusable = (*mockFocusable)(nil)
	})

	t.Run("Searchable", func(t *testing.T) {
		var _ Searchable = (*mockSearchable)(nil)
	})

	t.Run("Resizable", func(t *testing.T) {
		var _ Resizable = (*mockResizable)(nil)
	})

	t.Run("StatusProvider", func(t *testing.T) {
		var _ StatusProvider = (*mockStatusProvider)(nil)
	})
}

func TestInterfaces_MethodSignatures(t *testing.T) {
	t.Run("ComponentRenderer_Methods", func(t *testing.T) {
		typ := reflect.TypeOf((*ComponentRenderer)(nil)).Elem()

		// Check that interface has exactly 3 methods
		if typ.NumMethod() != 3 {
			t.Errorf("ComponentRenderer should have 3 methods, got %d", typ.NumMethod())
		}

		// Check method names exist
		expectedMethods := []string{"Init", "Update", "View"}
		for _, methodName := range expectedMethods {
			if _, ok := typ.MethodByName(methodName); !ok {
				t.Errorf("ComponentRenderer should have %s method", methodName)
			}
		}
	})

	t.Run("Focusable_Methods", func(t *testing.T) {
		typ := reflect.TypeOf((*Focusable)(nil)).Elem()

		expectedMethods := []string{"Focus", "Blur", "IsFocused"}
		for _, methodName := range expectedMethods {
			if _, ok := typ.MethodByName(methodName); !ok {
				t.Errorf("Focusable should have %s method", methodName)
			}
		}
	})

	t.Run("Searchable_Methods", func(t *testing.T) {
		typ := reflect.TypeOf((*Searchable)(nil)).Elem()

		expectedMethods := []string{"SetSearchQuery", "GetSearchQuery", "ClearSearch", "GetSearchResults"}
		for _, methodName := range expectedMethods {
			if _, ok := typ.MethodByName(methodName); !ok {
				t.Errorf("Searchable should have %s method", methodName)
			}
		}
	})
}

// Mock implementations for interface testing
type mockComponentRenderer struct{}

func (m *mockComponentRenderer) Init() tea.Cmd                           { return nil }
func (m *mockComponentRenderer) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *mockComponentRenderer) View() string                            { return "" }

type mockAppStateManager struct{}

func (m *mockAppStateManager) GetCurrentView() models.ViewType { return 0 }
func (m *mockAppStateManager) SetCurrentView(models.ViewType)  {}
func (m *mockAppStateManager) GetSelectedContext() string      { return "" }
func (m *mockAppStateManager) SetSelectedContext(string)       {}
func (m *mockAppStateManager) GetSelectedNamespace() string    { return "" }
func (m *mockAppStateManager) SetSelectedNamespace(string)     {}
func (m *mockAppStateManager) GetSelectedPod() *services.Pod   { return nil }
func (m *mockAppStateManager) SetSelectedPod(*services.Pod)    {}
func (m *mockAppStateManager) GetPodSearchQuery() string       { return "" }
func (m *mockAppStateManager) SetPodSearchQuery(string)        {}
func (m *mockAppStateManager) GetFilteredPods() []services.Pod { return nil }
func (m *mockAppStateManager) GetLogBuffer() []string          { return nil }
func (m *mockAppStateManager) IsFollowing() bool               { return false }
func (m *mockAppStateManager) SetFollowing(bool)               {}

type mockEventHandler struct{}

func (m *mockEventHandler) HandleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd)     { return nil, nil }
func (m *mockEventHandler) HandleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) { return nil, nil }
func (m *mockEventHandler) HandleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return nil, nil
}

type mockDataProvider struct{}

func (m *mockDataProvider) RefreshData() tea.Cmd  { return nil }
func (m *mockDataProvider) GetLoadingState() bool { return false }
func (m *mockDataProvider) GetErrorState() error  { return nil }

type mockFocusable struct{}

func (m *mockFocusable) Focus() tea.Cmd  { return nil }
func (m *mockFocusable) Blur() tea.Cmd   { return nil }
func (m *mockFocusable) IsFocused() bool { return false }

type mockSearchable struct{}

func (m *mockSearchable) SetSearchQuery(string)           {}
func (m *mockSearchable) GetSearchQuery() string          { return "" }
func (m *mockSearchable) ClearSearch()                    {}
func (m *mockSearchable) GetSearchResults() []interface{} { return nil }

type mockResizable struct{}

func (m *mockResizable) SetSize(width, height int)    {}
func (m *mockResizable) GetSize() (width, height int) { return 0, 0 }

type mockStatusProvider struct{}

func (m *mockStatusProvider) GetStatusText() string     { return "" }
func (m *mockStatusProvider) GetStatusType() StatusType { return StatusInfo }
