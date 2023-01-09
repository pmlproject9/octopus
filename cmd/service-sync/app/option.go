package app

import (
	"github.com/spf13/pflag"
)

type options struct {
	kubeconfig       string
	masterURL        string
	brokerKubeconfig string
	brokerMasterUrl  string
	clusterID        string
	namespace        string
}

func (opts *options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to a kubeconfig file for local cluster")
	fs.StringVar(&opts.masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig.")
	fs.StringVar(&opts.brokerKubeconfig, "brokerkubeconfig", "", "Path to a  kubeconfig file for broker cluster")
	fs.StringVar(&opts.brokerMasterUrl, "brokermaster", "", "The address of  the Kubernetes API server for the broker cluster. Overrides any value in kubeconfig.")
	fs.StringVar(&opts.clusterID, "clusterID", "", "The id of local cluster")
	fs.StringVar(&opts.namespace, "namespace", "", "")
}

func NewOptions() *options {
	return &options{}
}
