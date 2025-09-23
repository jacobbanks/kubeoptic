package services

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ConfigServiceImpl struct{}

func NewConfigService() ConfigService {
	return &ConfigServiceImpl{}
}

func (c *ConfigServiceImpl) DiscoverConfig() (string, error) {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		if _, err := os.Stat(kubeconfig); err == nil {
			return kubeconfig, nil
		}
	}

	// Check default location
	if home := homedir.HomeDir(); home != "" {
		configPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// TODO: Check in-cluster config
	return "", fmt.Errorf("no kubeconfig found")
}

func (c *ConfigServiceImpl) LoadContexts(configPath string) ([]Context, string, *kubernetes.Clientset, error) {
	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	// Parse kubeconfig
	kubeconfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Extract contexts
	contexts := make([]Context, 0, len(kubeconfig.Contexts))
	for name := range kubeconfig.Contexts {
		contexts = append(contexts, Context{Name: name})
	}

	// Get current context
	currentContext := kubeconfig.CurrentContext

	// Build REST config
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to build REST config: %w", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return contexts, currentContext, clientset, nil
}