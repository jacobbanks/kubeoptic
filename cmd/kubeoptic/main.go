package main

import (
	"fmt"
	"path/filepath"

	"flag"
	"kubeoptic/internal/k8s"
	// "kubeoptic/internal/models"

	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// kube "k8s.io/client-go/kubernetes"
	// cmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// TODO: Implement main entry point


func main() {
	var configpath *string
	if home := homedir.HomeDir(); home != "" {
		configpath = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to kube config file")
	} else {
		configpath = flag.String("kubeconfig", "", "absolute path to kube config")
	}

	flag.Parse()
	ko := k8s.Init(configpath)
	fmt.Printf("Current Context: %s, Available Context: %v", ko.SelectedContext, *ko.Contexts)
	// k8s.CountPods()
}
