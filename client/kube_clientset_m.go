package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
)

var KubeClientSetM kubeClientSetM

type kubeClientSetM struct {
	ClientSet *kubernetes.Clientset
}

type saJWTClaims struct {
	Sub string `json:"sub"` // "system:serviceaccount:dev:default"
}

func (k *kubeClientSetM) GetClient() *kubernetes.Clientset {
	return k.ClientSet
}

/*
Init kubernetes client
in-cluster > current workspace .kube/config > $HOME/.kube/config > flag kubeConfig file path
*/
func (k *kubeClientSetM) Init() {
	klog.Info("Starting init kube client...")
	var config *rest.Config
	// Try to use in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Warningf("Failed to get in-cluster kube config: %v, falling back to kubeConfig file", err)
		// Try to use config from current workspace directory
		configFile := filepath.Join(".", ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			klog.Warningf("Failed to load kube config from current workspace directory: %v, falling back to $HOME/.kube/config", err)
			// Try to use config from $HOME/.kube/config
			configFile = filepath.Join(homedir.HomeDir(), ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", configFile)
			if err != nil {
				klog.Warningf("Failed to load kube config from $HOME/.kube/config: %v, falling back to flag kubeConfig file", err)
				// Try to use config from flag kubeConfig file
				var kubeConfig *string
				kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
				flag.Parse()
				if *kubeConfig != "" {
					config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
					if err != nil {
						klog.Fatalf("Failed to load kube config from command line: %v", err)
					}
				} else {
					klog.Fatalf("Failed to load kube config from command line: kubeconfig flag is not set, exiting")
				}
			}
		}
	} else {
		// Check if the service account is "default"
		if config.BearerTokenFile != "" {
			err = k.checkServiceAccountIsDefault(config.BearerTokenFile)
			if err != nil {
				klog.Fatalf("Failed to check service account: %v, exiting", err)
			}
		}
	}

	klog.Info("Kube config loaded successfully, creating clientSet...")
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to get kube client: %v", err)
	}
	k.ClientSet = clientSet
}

// checkServiceAccount parses the JWT token to check if the service account is "default"
func (k *kubeClientSetM) checkServiceAccountIsDefault(tokenFile string) error {
	tokenBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		return errors.New("Failed to read service account token file: " + err.Error())
	}

	// Decode the JWT token
	tokenParts := strings.Split(string(tokenBytes), ".")
	if len(tokenParts) < 3 {
		return errors.New("invalid JWT token format")
	}

	// JWT payload is the second part, which is base64 encoded
	payload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return errors.New("Failed to decode JWT token payload: " + err.Error())
	}

	var claims saJWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return errors.New("Failed to unmarshal JWT claims: " + err.Error())
	}

	// The "sub" claim typically contains the format: system:serviceaccount:<namespace>:<serviceaccountname>
	subParts := strings.Split(claims.Sub, ":")
	if len(subParts) != 4 || subParts[1] != "serviceaccount" {
		return errors.New("unexpected JWT subject format")
	}

	serviceAccountName := subParts[3]
	if serviceAccountName == "default" {
		return errors.New("current service account is 'default', but it may not have enough permissions")
	}

	return nil
}
