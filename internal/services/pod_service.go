package services

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodServiceImpl struct {
	client kubernetes.Interface
}

func NewPodService(client kubernetes.Interface) PodService {
	return &PodServiceImpl{
		client: client,
	}
}

func (p *PodServiceImpl) ListPods(ctx context.Context, namespace string) ([]Pod, error) {
	podList, err := p.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	pods := make([]Pod, 0, len(podList.Items))
	for _, k8sPod := range podList.Items {
		pod := Pod{
			Name:      k8sPod.Name,
			Namespace: k8sPod.Namespace,
			Status:    convertPodStatus(k8sPod.Status.Phase),
			Labels:    k8sPod.Labels,
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (p *PodServiceImpl) SearchPods(ctx context.Context, namespace, query string) ([]Pod, error) {
	allPods, err := p.ListPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return allPods, nil
	}

	var filteredPods []Pod
	query = strings.ToLower(query)

	for _, pod := range allPods {
		if strings.Contains(strings.ToLower(pod.Name), query) {
			filteredPods = append(filteredPods, pod)
		}
	}

	return filteredPods, nil
}

func (p *PodServiceImpl) GetPodLogs(ctx context.Context, podName, namespace string) (io.ReadCloser, error) {
	req := p.client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs for pod %s/%s: %w", namespace, podName, err)
	}

	return stream, nil
}

func convertPodStatus(phase corev1.PodPhase) PodStatus {
	switch phase {
	case corev1.PodRunning:
		return PodRunning
	case corev1.PodPending:
		return PodPending
	case corev1.PodFailed:
		return PodFailed
	case corev1.PodSucceeded:
		return PodSucceeded
	default:
		return PodUnknown
	}
}