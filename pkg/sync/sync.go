package sync

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/submariner-io/admiral/pkg/syncer/broker"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sync"
)

type ServiceSyncConf struct {
	ClusterID  string
	RestMapper meta.RESTMapper
}

type Controller struct {
	clusterID     string
	kubeClientSet dynamic.Interface

	namespaceSyncer *broker.Syncer

	scheme     *runtime.Scheme
	restMapper meta.RESTMapper
}

func New(conf *ServiceSyncConf, kubeClient dynamic.Interface, brokerCfg *rest.Config) (*Controller, error) {
	var err error
	syncController := &Controller{
		clusterID:     conf.ClusterID,
		kubeClientSet: kubeClient,
		scheme:        scheme.Scheme,
		restMapper:    conf.RestMapper,
	}

	syncerConf := broker.SyncerConfig{}
	syncerConf.LocalNamespace = metav1.NamespaceAll
	syncerConf.LocalClusterID = conf.ClusterID
	syncerConf.RestMapper = conf.RestMapper
	syncerConf.LocalClient = kubeClient

	if brokerCfg != nil {
		syncerConf.BrokerClient, err = dynamic.NewForConfig(brokerCfg)
		if err != nil {
			return nil, errors.Wrap(err, "error to create dynamic client")
		}
	}

	syncerConf.ResourceConfigs = []broker.ResourceConfig{
		{
			LocalSourceNamespace: metav1.NamespaceAll,
			LocalResourceType:    &corev1.Namespace{},
			BrokerResourceType:   &corev1.Namespace{},
		},
	}
	syncController.namespaceSyncer, err = broker.NewSyncer(syncerConf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating namespace syncer")
	}

	return syncController, nil
}

func (ctr *Controller) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer utilruntime.HandleCrash()

	klog.Info("starting service syncer controller")

	if err := ctr.namespaceSyncer.Start(ctx.Done()); err != nil {
		klog.Errorf(fmt.Sprintf("%v", errors.Wrap(err, "error starting service syncer")))
	}
}
