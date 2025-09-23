package components

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/kubernetes"

	"kubeoptic/internal/models"
	"kubeoptic/internal/services"
	"kubeoptic/internal/tui"
)

// Integration tests for pod list component with K8s operations

// Mock services for integration testing
type mockPodServiceIntegration struct {
	pods     []services.Pod
	searchFn func(context.Context, string, string) ([]services.Pod, error)
}

func (m *mockPodServiceIntegration) ListPods(ctx context.Context, namespace string) ([]services.Pod, error) {
	return m.pods, nil
}

func (m *mockPodServiceIntegration) SearchPods(ctx context.Context, namespace, query string) ([]services.Pod, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, namespace, query)
	}

	// Default search implementation for testing
	var results []services.Pod
	queryLower := strings.ToLower(query)
	for _, pod := range m.pods {
		if strings.Contains(strings.ToLower(pod.Name), queryLower) ||
			strings.Contains(strings.ToLower(string(pod.Status)), queryLower) {
			results = append(results, pod)
		}
	}
	return results, nil
}

func (m *mockPodServiceIntegration) GetPodLogs(ctx context.Context, podName, namespace string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock log data")), nil
}

type mockNamespaceServiceIntegration struct {
	namespaces []string
}

func (m *mockNamespaceServiceIntegration) ListNamespaces(ctx context.Context) ([]string, error) {
	return m.namespaces, nil
}

func (m *mockNamespaceServiceIntegration) ListNamespacesDetailed(ctx context.Context) ([]services.Namespace, error) {
	var details []services.Namespace
	for _, ns := range m.namespaces {
		details = append(details, services.Namespace{
			Name:   ns,
			Status: "Active",
		})
	}
	return details, nil
}

type mockConfigServiceIntegration struct{}

func (m *mockConfigServiceIntegration) DiscoverConfig() (string, error) {
	return "/tmp/mock-kubeconfig", nil
}

func (m *mockConfigServiceIntegration) LoadContexts(configPath string) ([]services.Context, string, *kubernetes.Clientset, error) {
	contexts := []services.Context{
		{Name: "test-context"},
	}
	return contexts, "test-context", nil, nil
}

// Integration test data
var integrationTestPods = []services.Pod{
	{
		Name:      "web-server-123",
		Namespace: "production",
		Status:    services.PodRunning,
		Labels:    map[string]string{"app": "web", "tier": "frontend"},
	},
	{
		Name:      "database-456",
		Namespace: "production",
		Status:    services.PodRunning,
		Labels:    map[string]string{"app": "db", "tier": "backend"},
	},
	{
		Name:      "worker-789",
		Namespace: "production",
		Status:    services.PodPending,
		Labels:    map[string]string{"app": "worker", "tier": "background"},
	},
	{
		Name:      "failed-job-001",
		Namespace: "jobs",
		Status:    services.PodFailed,
		Labels:    map[string]string{"job": "cleanup"},
	},
}

func TestPodListIntegration(t *testing.T) {
	// Setup mock services
	podSvc := &mockPodServiceIntegration{pods: integrationTestPods}
	namespaceSvc := &mockNamespaceServiceIntegration{namespaces: []string{"production", "jobs"}}
	configSvc := &mockConfigServiceIntegration{}

	// Create kubeoptic model
	kubeoptic := models.NewKubeoptic(configSvc, podSvc, namespaceSvc)

	t.Run("end_to_end_pod_selection", func(t *testing.T) {
		// Create pod list component
		podList := NewPodList(integrationTestPods, 100, 30)
		podList.Focus()

		// Set namespace so kubeoptic can search for pods
		err := kubeoptic.SelectNamespace("production")
		if err != nil {
			t.Errorf("Failed to select namespace: %v", err)
		}

		// Simulate user selecting first pod
		_, cmd := podList.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("Expected command from pod selection")
		}

		// Execute command to get the message
		msg := cmd()
		if podSelectedMsg, ok := msg.(tui.PodSelectedMsg); ok {
			if podSelectedMsg.Pod == nil {
				t.Fatal("Selected pod should not be nil")
			}

			// Verify correct pod was selected (first item)
			expectedPod := integrationTestPods[0]
			if podSelectedMsg.Pod.Name != expectedPod.Name {
				t.Errorf("Expected pod %s, got %s", expectedPod.Name, podSelectedMsg.Pod.Name)
			}

			// For this test, we'll verify the pod selection message is correct
			// In a real application, kubeoptic would handle this selection
			if podSelectedMsg.Pod.Namespace != expectedPod.Namespace {
				t.Errorf("Expected namespace %s, got %s", expectedPod.Namespace, podSelectedMsg.Pod.Namespace)
			}

			if podSelectedMsg.Pod.Status != expectedPod.Status {
				t.Errorf("Expected status %s, got %s", expectedPod.Status, podSelectedMsg.Pod.Status)
			}
		} else {
			t.Errorf("Expected PodSelectedMsg, got %T", msg)
		}
	})

	t.Run("search_integration", func(t *testing.T) {
		// Setup search function for mock service
		podSvc.searchFn = func(ctx context.Context, namespace, query string) ([]services.Pod, error) {
			var results []services.Pod
			queryLower := strings.ToLower(query)
			for _, pod := range integrationTestPods {
				if strings.Contains(strings.ToLower(pod.Name), queryLower) {
					results = append(results, pod)
				}
			}
			return results, nil
		}

		// Create pod list
		searchPodList := NewPodList(integrationTestPods, 100, 30)

		// Simulate search activation
		_, cmd := searchPodList.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		if cmd != nil {
			t.Error("Search activation should not return command")
		}

		if !searchPodList.searching {
			t.Error("Search should be activated")
		}

		// Test search query message handling
		searchMsg := tui.SearchQueryChangedMsg{Query: "web"}
		model, cmd := searchPodList.Update(searchMsg)
		if cmd != nil {
			t.Error("Search query change should not return command")
		}

		updatedList := model.(*PodList)
		if !updatedList.searching {
			t.Error("Search should remain active after query change")
		}

		// Simulate kubeoptic search
		err := kubeoptic.SearchPods("web")
		if err != nil {
			t.Errorf("Kubeoptic search failed: %v", err)
		}

		// Verify search results would be filtered (this would come from kubeoptic)
		filteredPods := kubeoptic.GetPods()
		if len(filteredPods) == 0 {
			t.Error("Search should return filtered results")
		}

		// Update pod list with search results
		searchResultsMsg := tui.PodsLoadedMsg{Pods: filteredPods, Error: nil}
		_, cmd = updatedList.Update(searchResultsMsg)
		if cmd != nil {
			t.Error("Pods loaded message should not return command")
		}
	})

	t.Run("pod_status_filtering", func(t *testing.T) {
		// Test that all status types are represented
		statusCounts := make(map[services.PodStatus]int)
		for _, pod := range integrationTestPods {
			statusCounts[pod.Status]++
		}

		if statusCounts[services.PodRunning] == 0 {
			t.Error("Should have running pods in test data")
		}
		if statusCounts[services.PodPending] == 0 {
			t.Error("Should have pending pods in test data")
		}
		if statusCounts[services.PodFailed] == 0 {
			t.Error("Should have failed pods in test data")
		}

		// Verify component can handle status-based filtering
		err := kubeoptic.SearchPods("Running")
		if err != nil {
			t.Errorf("Status-based search failed: %v", err)
		}

		runningPods := kubeoptic.GetPods()
		for _, pod := range runningPods {
			if pod.Status != services.PodRunning {
				t.Error("Status search should only return running pods")
			}
		}
	})

	t.Run("namespace_context_integration", func(t *testing.T) {
		// Test different namespaces
		productionPods := []services.Pod{}
		jobsPods := []services.Pod{}

		for _, pod := range integrationTestPods {
			if pod.Namespace == "production" {
				productionPods = append(productionPods, pod)
			} else if pod.Namespace == "jobs" {
				jobsPods = append(jobsPods, pod)
			}
		}

		// Create pod list for production namespace
		prodPodList := NewPodList(productionPods, 100, 30)
		if len(prodPodList.pods) != len(productionPods) {
			t.Error("Production pod list should have correct number of pods")
		}

		// Create pod list for jobs namespace
		jobsPodList := NewPodList(jobsPods, 100, 30)
		if len(jobsPodList.pods) != len(jobsPods) {
			t.Error("Jobs pod list should have correct number of pods")
		}

		// Verify namespace isolation
		if len(prodPodList.pods) == len(jobsPodList.pods) {
			t.Error("Different namespaces should have different pod counts")
		}
	})

	t.Run("error_handling_integration", func(t *testing.T) {
		// Test error handling in pod list
		podList := NewPodList(integrationTestPods, 100, 30)

		// Simulate error from kubeoptic
		errorMsg := tui.PodsLoadedMsg{
			Pods:  nil,
			Error: &testErrorIntegration{message: "network timeout"},
		}

		model, cmd := podList.Update(errorMsg)
		if cmd != nil {
			t.Error("Error message should not return command")
		}

		updatedList := model.(*PodList)
		if updatedList == nil {
			t.Error("Component should handle errors gracefully")
		}

		// Verify component state is preserved during error
		if len(updatedList.pods) != len(integrationTestPods) {
			t.Error("Pod list should preserve previous state on error")
		}
	})

	t.Run("responsive_layout_integration", func(t *testing.T) {
		// Test responsive behavior with different sizes
		testSizes := []struct {
			width, height int
			name          string
		}{
			{80, 20, "standard"},
			{120, 40, "large"},
			{60, 15, "small"},
			{200, 50, "extra_large"},
		}

		for _, size := range testSizes {
			t.Run(size.name, func(t *testing.T) {
				podList := NewPodList(integrationTestPods, size.width, size.height)

				// Test resize
				resizeMsg := tea.WindowSizeMsg{Width: size.width + 20, Height: size.height + 10}
				model, cmd := podList.Update(resizeMsg)
				if cmd != nil {
					t.Error("Resize should not return command")
				}

				updatedList := model.(*PodList)
				w, h := updatedList.GetSize()
				if w != size.width+20 || h != size.height+10 {
					t.Errorf("Size not updated correctly: expected %dx%d, got %dx%d",
						size.width+20, size.height+10, w, h)
				}

				// Verify component still renders at any size
				view := updatedList.View()
				if view == "" {
					t.Error("Component should render at any size")
				}
			})
		}
	})
}

func TestPodListPerformance(t *testing.T) {
	// Performance test with large dataset
	largePodList := make([]services.Pod, 1000)
	for i := 0; i < 1000; i++ {
		largePodList[i] = services.Pod{
			Name:      fmt.Sprintf("pod-%d", i),
			Namespace: fmt.Sprintf("namespace-%d", i%10),
			Status: services.PodStatus([]services.PodStatus{
				services.PodRunning,
				services.PodPending,
				services.PodFailed,
			}[i%3]),
			Labels: map[string]string{
				"app":  fmt.Sprintf("app-%d", i%5),
				"tier": fmt.Sprintf("tier-%d", i%3),
			},
		}
	}

	t.Run("large_dataset_creation", func(t *testing.T) {
		// Should handle large datasets without performance issues
		podList := NewPodList(largePodList, 100, 30)
		if podList == nil {
			t.Fatal("Should create pod list with large dataset")
		}

		if len(podList.pods) != 1000 {
			t.Errorf("Expected 1000 pods, got %d", len(podList.pods))
		}
	})

	t.Run("large_dataset_updates", func(t *testing.T) {
		podList := NewPodList(largePodList, 100, 30)

		// Test update performance
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		model, cmd := podList.Update(msg)
		if cmd != nil {
			t.Error("Update should not return command")
		}

		if model == nil {
			t.Error("Update should return valid model")
		}
	})

	t.Run("large_dataset_rendering", func(t *testing.T) {
		podList := NewPodList(largePodList, 100, 30)
		podList.Focus()

		// Test rendering performance
		view := podList.View()
		if view == "" {
			t.Error("Should render large dataset")
		}

		// Verify rendering doesn't crash with large datasets
		if len(view) == 0 {
			t.Error("Rendered view should have content")
		}
	})
}

// Helper types for integration testing
type testErrorIntegration struct {
	message string
}

func (e *testErrorIntegration) Error() string {
	return e.message
}

// Benchmark integration test
func BenchmarkPodListIntegration(b *testing.B) {
	podList := NewPodList(integrationTestPods, 100, 30)
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		podList.Update(msg)
	}
}
