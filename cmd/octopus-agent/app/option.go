package app

import (
	"github.com/spf13/pflag"
	"time"
)

type options struct {
	kubeconfig         string
	masterURL          string
	IPTablesSyncPeriod time.Duration
	clusterID          string
	brokerKubeconfig   string
	brokerMasterUrl    string
}

func (opts *options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to a kubeconfig file for local cluster")
	fs.StringVar(&opts.masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig.")
	fs.DurationVar(&opts.IPTablesSyncPeriod, "iptables-sync-period", time.Second*5, "The delay between iptables rule synchronizations (e.g. '5s', '1m'). Must be greater than 0.")
	fs.StringVar(&opts.clusterID, "clusterID", "", "The  ID of local cluster")
	fs.StringVar(&opts.brokerKubeconfig, "brokerkubeconfig", "", "Path to a  kubeconfig file for broker cluster")
	fs.StringVar(&opts.brokerMasterUrl, "brokermaster", "", "The address of  the Kubernetes API server for the broker cluster. Overrides any value in kubeconfig.")
}

func NewOptions() *options {
	return &options{}
}
