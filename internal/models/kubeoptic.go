package models 

import (
	"context"
	"fmt"
	"bufio"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)	

type KubeOptic struct {
	client *kubernetes.Clientset
	Contexts *[]string
	KubeConfigPath string
	SelectedContext string
	SelectedSubscription string
	selectedPod string
	Subscriptions *[]string
	RawConfig *[]byte
}




func (k *KubeOptic) AttachClient(c *kubernetes.Clientset) {
	k.client = c
}


func (k *KubeOptic) PrintLogsForPod() {
	req := k.client.CoreV1().Pods("default").GetLogs(k.selectedPod, &corev1.PodLogOptions{}) 
	reader, err := req.Stream(context.TODO())
	if err != nil {
		panic(err.Error())
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log := scanner.Text()
		fmt.Printf("%s \n", log)
	}
}

func (k *KubeOptic) ListPods() {
	pods, err := k.client.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		fmt.Printf("Pod Name: %s \n", pod.Name)
	}
}


func (k *KubeOptic) SelectPod(podName string) {
	pods, err := k.client.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		if pod.Name == podName {
			k.selectedPod = pod.Name
		}
	}
	fmt.Printf("Selected Pod: %s", k.selectedPod)
}

func (k *KubeOptic) SelectContext(newContext string) {
	k.SelectedContext = newContext
}

func (k *KubeOptic) SelectSubscription(subscription string) {
	k.SelectedSubscription = subscription
}

func (k *KubeOptic) SetKubeConfigPath(path string) {
	k.KubeConfigPath = path
}

func (k *KubeOptic) SetSubscriptions(subs *[]string) {
	k.Subscriptions = subs
}

func (k *KubeOptic) SetContexts(contexts *[]string) {
	k.Contexts = contexts
}
