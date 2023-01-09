package app

import (
	"context"
	octopusv1alpha1 "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	"github.com/pmlproject9/octopus/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/submariner-io/admiral/pkg/util"
	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/yaml"
)

var (
	cmdName = `service-syncer`
)

func NewServiceSyncCmd(ctx context.Context) *cobra.Command {
	opts := NewOptions()
	cmd := &cobra.Command{
		Use:  cmdName,
		Long: `Responsible for sync service of all cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			var cfg *rest.Config
			var brokerCfg *rest.Config
			var err error

			err = octopusv1alpha1.AddToScheme(scheme.Scheme)
			if err != nil {
				klog.Fatalf("Error adding octopus v1alpha1 to the scheme:%v", err)
			}

			if opts.masterURL != "" || opts.kubeconfig != " " {
				cfg, err = clientcmd.BuildConfigFromFlags(opts.masterURL, opts.kubeconfig)
				if err != nil {
					klog.Fatalf("Error build kubeconfig: %+v", err)
				}
			} else {
				cfg, err = rest.InClusterConfig()
				if err != nil {
					klog.Fatalf("unable to initialize inclusterconfig: %+v", err)
				}
			}
			if opts.brokerKubeconfig != "" || opts.brokerMasterUrl != "" {
				brokerCfg, err = clientcmd.BuildConfigFromFlags(opts.brokerMasterUrl, opts.brokerKubeconfig)
				if err != nil {
					klog.Fatalf("Error build kubeconfig: %+v", err)
				}
			}

			localClient, err := dynamic.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("error creating dynamic client: %v", err.Error())
			}

			restMapper, err := util.BuildRestMapper(cfg)
			if err != nil {
				klog.Fatalf("error building rest mapper:%v", err.Error())
			}
			conf := &sync.ServiceSyncConf{
				ClusterID:  opts.clusterID,
				Namespace:  opts.namespace,
				RestMapper: restMapper,
			}

			syncController, err := sync.New(conf, restMapper, localClient, brokerCfg)
			if err != nil {
				klog.Fatalf("Failed to create servicesync:%v", err)
			}

			cmCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			if err := syncController.Run(cmCtx); err != nil {
				klog.Fatalf("Failed to start servicesync: %v", err)
			}

			<-ctx.Done()

		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

func readFileToCrd(path string, obj *extensionv1.CustomResourceDefinition) {
	jsonByte, err := os.ReadFile(path)
	if err != nil {
		klog.Fatalf("error read file %s:%v ", path, err.Error())
	}
	err = yaml.Unmarshal(jsonByte, obj)
	if err != nil {
		klog.Fatalf("error read file %s:%v ", path, err.Error())
	}
}
