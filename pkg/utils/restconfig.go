package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type tokenConfig struct {
	Apiserver      string
	ApiserverToken string
	Ca             string
	Insecure       bool
}

func GenerateConfigFromEnvironment() *rest.Config {
	cfg := tokenConfig{}
	err := envconfig.Process("broker_k8s", &cfg)
	if err != nil {
		klog.Fatal("error to read broker kube config")
	}
	fmt.Println(cfg.Apiserver)
	fmt.Println(cfg.ApiserverToken)
	tls := &rest.TLSClientConfig{}
	if !cfg.Insecure && cfg.Ca != "" {
		caDecoded, err := base64.StdEncoding.DecodeString(cfg.Ca)
		if err != nil {
			klog.Fatalf("error to broker kubeconfig config : %v", err)
		}
		tls.Insecure = false
		tls.CAData = caDecoded
	}

	return &rest.Config{
		Host:            "https://" + cfg.Apiserver,
		BearerToken:     cfg.ApiserverToken,
		TLSClientConfig: *tls,
	}
}
