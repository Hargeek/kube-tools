package client

import (
	"errors"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
)

var Helm helm

type helm struct {
	ClientMap   map[string]*action.Configuration // multi-cluster client
	HelmConfMap map[string]string                // cluster config
}

// GetClient get helm client by cluster name
func (h *helm) GetClient(cluster string) (*action.Configuration, error) {
	client, ok := h.ClientMap[cluster]
	if !ok {
		return nil, errors.New(fmt.Sprintf("cluster:%s not exist, can't get client\n", cluster))
	}
	return client, nil
}

// Init helm client
func (h *helm) Init() {}
