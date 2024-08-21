package main

import "github.com/Hargeek/kube-tools/client"

func main() {
	client.KubeClientSetWithConfig.Init()
	//client.Helm.Init()
}
