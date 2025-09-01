package k8s

import (
	"kubeoptic/internal/models"
	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	cmd "k8s.io/client-go/tools/clientcmd"
	"os"
)

func loadDataFromConfig(ko *models.KubeOptic) {
	kubeconfig, err := cmd.LoadFromFile(ko.KubeConfigPath)
	if err != nil {
		panic(err.Error())
	}

	contexts := make([]string, 0, len(kubeconfig.Contexts))
	for name := range kubeconfig.Contexts {
		contexts = append(contexts, name)
	}
	ko.SetContexts(&contexts)
	ko.SelectContext(kubeconfig.CurrentContext)
	restConfig, err := cmd.RESTConfigFromKubeConfig(*ko.RawConfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}
	ko.AttachClient(clientset)
	ko.ListPods()
	ko.SelectPod("api2-projects-496sportcalcs-5d89bd45b6-7xs96")
	ko.PrintLogsForPod()
}

func Init(configPath *string) models.KubeOptic {
	byteData, err := os.ReadFile(*configPath)
	if err != nil {
		panic(err.Error())
	}

	k := &models.KubeOptic{
		KubeConfigPath: *configPath,
		RawConfig: &byteData,
	}
	loadDataFromConfig(k)
	return *k
}

