package k8s

// TODO: Implement Kubernetes client wrapper
import (
	"context"
	"fmt"
	"path/filepath"

	"flag"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func CountPods() {
	var config *string
	if home := homedir.HomeDir(); home != "" {
		config = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to kube config file")
	} else {
		config = flag.String("kubeconfig", "", "absolute path to kube config")
	}
	flag.Parse()

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", *config)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		panic(err.Error())
	}


	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster", len(pods.Items))

}


