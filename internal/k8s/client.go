package k8s

// TODO: Implement Kubernetes client wrapper
import (
	"fmt"
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getConfig()
	var config *string
	if home := homedir.HomeDir(); home !== "" {
	config = flag.string("config")
}



