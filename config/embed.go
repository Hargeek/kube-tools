package config

import (
	"embed"
)

//go:embed .kube/*
var KubeFS embed.FS

func GetKubeEmbed() embed.FS {
	return KubeFS
}
