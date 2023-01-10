package controllermanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/pmlproject9/octopus/pkg/controller/multinetworkpolicy"
	"github.com/submariner-io/admiral/pkg/util"
	"k8s.io/client-go/dynamic"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	octopusclientset "github.com/pmlproject9/octopus/pkg/generated/clientset/versioned"
	octopusinformers "github.com/pmlproject9/octopus/pkg/generated/informers/externalversions"
	syncCtr "github.com/pmlproject9/octopus/pkg/sync"
	submarinerClientset "github.com/submariner-io/submariner/pkg/client/clientset/versioned"
	submarinerinformers "github.com/submariner-io/submariner/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	mcsClientset "sigs.k8s.io/mcs-api/pkg/client/clientset/versioned"
	mcsInformers "sigs.k8s.io/mcs-api/pkg/client/informers/externalversions"
)

var DefaultResync = time.Hour * 12

type ControllerManager struct {
	ctx context.Context

	k8sClient kubernetes.Interface
	mcsClient mcsClientset.Interface

	kubeInformerFactory       kubeinformers.SharedInformerFactory
	octopusInformerFactory    octopusinformers.SharedInformerFactory
	submarinerInformerFactory submarinerinformers.SharedInformerFactory
	mcsInformerFactory        mcsInformers.SharedInformerFactory

	mnpController  *multinetworkpolicy.Controller
	syncController *syncCtr.Controller
}

func NewControllerManager(ctx context.Context,
	cfg *rest.Config,
	brokerCfg *rest.Config,
	syncPeriod time.Duration,
	clusterID string) (*ControllerManager, error) {

	k8sClinet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate kubernetes client: %+v", err))
	}

	restMapper, err := util.BuildRestMapper(cfg)
	if err != nil {
		klog.Fatalf("error building rest mapper:%v", err.Error())
	}
	conf := &syncCtr.ServiceSyncConf{
		ClusterID:  clusterID,
		RestMapper: restMapper,
	}

	localClient := dynamic.NewForConfigOrDie(cfg)

	syncController, err := syncCtr.New(conf, localClient, brokerCfg)
	if err != nil {
		klog.Fatalf("Failed to create servicesync:%v", err)
	}

	octopusBrokerclient, err := octopusclientset.NewForConfig(brokerCfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate octopus broker client: %+v", err))
	}

	submarinerClient, err := submarinerClientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate submairner client: %+v", err))
	}
	mcsClient, err := mcsClientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error to generate mcs client: %+v", err))
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(k8sClinet, DefaultResync)
	submarinerInformerFactory := submarinerinformers.NewSharedInformerFactory(submarinerClient, DefaultResync)
	mcsInformerFactory := mcsInformers.NewSharedInformerFactory(mcsClient, DefaultResync)

	octopusInformerFactory := octopusinformers.NewSharedInformerFactory(octopusBrokerclient, DefaultResync)

	namespaceInformer := kubeInformerFactory.Core().V1().Namespaces().Informer()
	podInformer := kubeInformerFactory.Core().V1().Pods().Informer()
	multiNetworkPolicyInformer := octopusInformerFactory.Octopus().V1alpha1().MultiNetworkPolicies().Informer()
	endpointsInformer := submarinerInformerFactory.Submariner().V1().Endpoints().Informer()
	serviceImportInformer := mcsInformerFactory.Multicluster().V1alpha1().ServiceImports().Informer()

	mnpController := multinetworkpolicy.NewMultiNetworkPolicyController(ctx, syncPeriod, k8sClinet, mcsClient, namespaceInformer,
		podInformer, multiNetworkPolicyInformer, serviceImportInformer, endpointsInformer, clusterID)

	return &ControllerManager{
		ctx:                       ctx,
		k8sClient:                 k8sClinet,
		mcsClient:                 mcsClient,
		kubeInformerFactory:       kubeInformerFactory,
		octopusInformerFactory:    octopusInformerFactory,
		submarinerInformerFactory: submarinerInformerFactory,
		mcsInformerFactory:        mcsInformerFactory,
		mnpController:             mnpController,
		syncController:            syncController,
	}, nil

}

func (cm *ControllerManager) Run(ctx context.Context) {
	klog.Info("starting controller manager running....")

	cm.kubeInformerFactory.Start(ctx.Done())
	cm.octopusInformerFactory.Start(ctx.Done())
	cm.submarinerInformerFactory.Start(ctx.Done())
	cm.mcsInformerFactory.Start(ctx.Done())

	var wg sync.WaitGroup
	wg.Add(2)
	go cm.mnpController.Run(&wg)
	go cm.syncController.Run(cm.ctx, &wg)
	wg.Wait()

}
