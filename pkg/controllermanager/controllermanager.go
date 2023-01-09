package controllermanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/pmlproject9/octopus/pkg/controller/multinetworkpolicy"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	octopusclientset "github.com/pmlproject9/octopus/pkg/generated/clientset/versioned"
	octopusinformers "github.com/pmlproject9/octopus/pkg/generated/informers/externalversions"
	submarinerClientset "github.com/submariner-io/submariner/pkg/client/clientset/versioned"
	submarinerinformers "github.com/submariner-io/submariner/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
)

var DefaultResync = time.Hour * 12

type ControllerManager struct {
	ctx context.Context

	k8sClient     kubernetes.Interface
	octopusClient *octopusclientset.Clientset

	kubeInformerFactory       kubeinformers.SharedInformerFactory
	octopusInformerFactory    octopusinformers.SharedInformerFactory
	submarinerInformerFactory submarinerinformers.SharedInformerFactory
	mnpController             *multinetworkpolicy.Controller
}

func NewControllerManager(ctx context.Context, cfg *rest.Config, syncPeriod time.Duration, clusterID string) (*ControllerManager, error) {

	k8sClinet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate kubernetes client: %+v", err))
	}
	octopusclient, err := octopusclientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate octopus client: %+v", err))
	}

	submarinerClient, err := submarinerClientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate submairner client: %+v", err))
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(k8sClinet, DefaultResync)
	octopusInformerFactory := octopusinformers.NewSharedInformerFactory(octopusclient, DefaultResync)
	submarinerInformerFactory := submarinerinformers.NewSharedInformerFactory(submarinerClient, DefaultResync)

	namespaceInformer := kubeInformerFactory.Core().V1().Namespaces().Informer()
	podInformer := kubeInformerFactory.Core().V1().Pods().Informer()
	multiNetworkPolicyInformer := octopusInformerFactory.Octopus().V1alpha1().MultiNetworkPolicies().Informer()
	servicesyncInformer := octopusInformerFactory.Octopus().V1alpha1().ServiceSyncs().Informer()
	endpointsInformer := submarinerInformerFactory.Submariner().V1().Endpoints().Informer()

	mnpController := multinetworkpolicy.NewMultiNetworkPolicyController(ctx, syncPeriod, k8sClinet, namespaceInformer,
		podInformer, multiNetworkPolicyInformer, servicesyncInformer, endpointsInformer, clusterID)

	return &ControllerManager{
		ctx:                       ctx,
		k8sClient:                 k8sClinet,
		octopusClient:             octopusclient,
		kubeInformerFactory:       kubeInformerFactory,
		octopusInformerFactory:    octopusInformerFactory,
		submarinerInformerFactory: submarinerInformerFactory,
		mnpController:             mnpController,
	}, nil

}

func (cm *ControllerManager) Run(ctx context.Context) {
	klog.Info("starting controller manager running....")

	cm.kubeInformerFactory.Start(ctx.Done())
	cm.octopusInformerFactory.Start(ctx.Done())
	cm.submarinerInformerFactory.Start(ctx.Done())

	var wg sync.WaitGroup
	wg.Add(1)
	go cm.mnpController.Run(&wg)
	wg.Wait()

}
