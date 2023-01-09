package app

import (
	"context"
	"github.com/pmlproject9/octopus/pkg/controllermanager"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	cmdName = "octopus-agent"
)

func NewOctopusAgent(ctx context.Context) *cobra.Command {
	opts := NewOptions()
	cmd := &cobra.Command{
		Use:  cmdName,
		Long: `Responsible for network policy of multi cluster service`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			var cfg *rest.Config
			var err error

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

			cmCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			cm, err := controllermanager.NewControllerManager(cmCtx, cfg, opts.IPTablesSyncPeriod, opts.clusterID)
			if err != nil {
				klog.Fatalf("Error run controller manager: %+v", err)
			}
			cm.Run(cmCtx)
		},
	}

	opts.AddFlags(cmd.Flags())
	return cmd
}
