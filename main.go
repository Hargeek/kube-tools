package main

import "github.com/Hargeek/kube-tools/client"

func main() {
	client.K8sCS.Init()
	//client.Helm.Init()
}
