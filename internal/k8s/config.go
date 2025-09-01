package k8s

// TODO: Implement kubeconfig handling
import (
	// "context"
	"flag"
	"fmt"
	"path/filepath"

	// "github.com/emicklei/go-restful/v3/log"
	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "github.com/google/go-cmp/cmp/internal/flags"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ConfigSummary struct {
	Host	string `json:"Host"`
}

func GetConfig() {
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

	fmt.Println("Error", kubeconfig.GoString())

	// fmt.Printf("There are %v pods in the cluster", kubeconfig)

}


