package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Hargeek/kube-tools/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var K8sCS k8sCS

type k8sCS struct {
	ClientMap   map[string]*kubernetes.Clientset // multi-cluster client
	KubeConfMap map[string]string                // cluster config
}

// GetClient get k8s client by cluster name
func (k *k8sCS) GetClient(cluster string) (*kubernetes.Clientset, error) {
	client, ok := k.ClientMap[cluster]
	if !ok {
		return nil, errors.New(fmt.Sprintf("cluster:%s not exist, can't get client\n", cluster))
	}
	return client, nil
}

// Init k8s client
func (k *k8sCS) Init() {
	mp := map[string]string{}
	k.ClientMap = map[string]*kubernetes.Clientset{}
	if err := json.Unmarshal([]byte(config.KubeConfigRelativePath), &mp); err != nil {
		klog.Fatalf("kube config init unmarshal failed %v\n", err)
	}
	k.KubeConfMap = mp
	for clusterName, kubeConfigFilePath := range mp {
		kubeConfigFileData, err := config.GetKubeEmbed().ReadFile(kubeConfigFilePath)
		if err != nil {
			klog.Fatalf("cluster %s: read kubeconfig file failed %v\n", clusterName, err)
		}
		//conf, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
		conf, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigFileData)
		if err != nil {
			klog.Fatalf("cluster %s: get k8s config failed %v\n", clusterName, err)
		}
		clientSet, err := kubernetes.NewForConfig(conf)
		if err != nil {
			klog.Fatalf("cluster %s: init k8s client failed %v\n", clusterName, err)
		}
		k.ClientMap[clusterName] = clientSet
		klog.Infof("cluster %s: init k8s client success ", clusterName)
	}
}
