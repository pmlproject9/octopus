package sync

import (
	"context"
	"github.com/pkg/errors"
	octopusV1alpha1 "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	"github.com/pmlproject9/octopus/pkg/constants"
	"github.com/submariner-io/admiral/pkg/log"
	"github.com/submariner-io/admiral/pkg/syncer"
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
)

type ServiceSyncConf struct {
	ClusterID  string
	Namespace  string
	RestMapper meta.RESTMapper
}

type Controller struct {
	clusterID         string
	namespace         string
	kubeClientSet     dynamic.Interface
	serviceSyncSyncer *broker.Syncer
	mnpSyncer         *broker.Syncer
	namespaceSyncer   *broker.Syncer
	serviceSyncer     syncer.Interface

	scheme            *runtime.Scheme
	restMapper        meta.RESTMapper
	serviceSyncClient dynamic.NamespaceableResourceInterface
}

func New(conf *ServiceSyncConf, restMapper meta.RESTMapper, kubeClient dynamic.Interface, brokerCfg *rest.Config) (*Controller, error) {
	var err error
	syncController := &Controller{
		clusterID:     conf.ClusterID,
		namespace:     conf.Namespace,
		kubeClientSet: kubeClient,
		scheme:        scheme.Scheme,
		restMapper:    restMapper,
	}

	syncerConf := broker.SyncerConfig{}
	syncerConf.LocalNamespace = metav1.NamespaceAll
	syncerConf.LocalClusterID = conf.ClusterID
	syncerConf.RestMapper = restMapper
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
			LocalResourceType:    &octopusV1alpha1.ServiceSync{},
			BrokerResourceType:   &octopusV1alpha1.ServiceSync{},
		},
	}
	syncController.serviceSyncSyncer, err = broker.NewSyncer(syncerConf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating servicesync syncer")
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

	syncerConf.ResourceConfigs = []broker.ResourceConfig{
		{
			LocalSourceNamespace: metav1.NamespaceAll,
			LocalResourceType:    &octopusV1alpha1.MultiNetworkPolicy{},
			BrokerResourceType:   &octopusV1alpha1.MultiNetworkPolicy{},
		},
	}
	syncController.mnpSyncer, err = broker.NewSyncer(syncerConf)
	if err != nil {
		return nil, errors.Wrap(err, "error creating multinetworkpolicy syncer")
	}

	syncController.serviceSyncer, err = syncer.NewResourceSyncer(&syncer.ResourceSyncerConfig{
		Name:            "Service -> ServiceSync",
		SourceClient:    kubeClient,
		SourceNamespace: metav1.NamespaceAll,
		RestMapper:      conf.RestMapper,
		Federator:       syncController.serviceSyncSyncer.GetLocalFederator(),
		ResourceType:    &corev1.Service{},
		Transform:       syncController.serviceToServiceSync,
		Scheme:          syncController.scheme,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating service syncer")
	}

	return syncController, nil

}

func (ctr *Controller) Run(ctx context.Context) error {
	defer utilruntime.HandleCrash()

	klog.Info("starting service syncer controller")

	if err := ctr.serviceSyncer.Start(ctx.Done()); err != nil {
		return errors.Wrap(err, "error starting service syncer")
	}
	if err := ctr.mnpSyncer.Start(ctx.Done()); err != nil {
		return errors.Wrap(err, "error starting multinetworkpolicy syncer")
	}
	if err := ctr.serviceSyncSyncer.Start(ctx.Done()); err != nil {
		return errors.Wrap(err, "error starting servicesync syncer")
	}

	if err := ctr.namespaceSyncer.Start(ctx.Done()); err != nil {
		return errors.Wrap(err, "error to start namespace syncer")
	}
	ctr.serviceSyncer.Reconcile(func() []runtime.Object {
		return ctr.serviceSyncLister(func(ss *octopusV1alpha1.ServiceSync) runtime.Object {
			return &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ss.GetLabels()[constants.LabelSourceName],
					Namespace: ss.GetLabels()[constants.LabelSourceNamespace],
				},
			}
		})
	})
	return nil

}

func (ctr *Controller) serviceToServiceSync(obj runtime.Object, numRequeues int, op syncer.Operation) (runtime.Object, bool) {
	svc := obj.(*corev1.Service)

	klog.V(log.DEBUG).Infof("service %s/%s %s", svc.Namespace, svc.Name, op)

	if op == syncer.Delete {
		return ctr.newServiceSync(svc.Name, svc.Namespace), false
	}

	serviceSync := ctr.newServiceSync(svc.Name, svc.Namespace)

	for labelKey, value := range svc.Labels {
		serviceSync.Labels[labelKey] = value
	}

	serviceSync.Spec = octopusV1alpha1.ServiceSyncSpec{
		IPs:            svc.Spec.ClusterIPs,
		IPFamilies:     svc.Spec.IPFamilies,
		IPFamilyPolicy: svc.Spec.IPFamilyPolicy,
	}
	serviceSync.Spec.Ports = ctr.getPorts(svc)

	klog.V(log.DEBUG).Infof("Returning ServiceSync: %#v", serviceSync)
	return serviceSync, false

}

func (ctr *Controller) newServiceSync(name, namespace string) *octopusV1alpha1.ServiceSync {
	return &octopusV1alpha1.ServiceSync{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctr.getObjectNameWithClusterID(name, namespace),
			Namespace: namespace,
			Labels: map[string]string{
				constants.LabelSourceNamespace: namespace,
				constants.LabelSourceName:      name,
				constants.LabelSourceClusterID: ctr.clusterID,
			},
		},
	}
}

func (ctr *Controller) getPorts(svc *corev1.Service) []octopusV1alpha1.ServicePort {
	ports := make([]octopusV1alpha1.ServicePort, 0)
	for _, port := range svc.Spec.Ports {
		ports = append(ports, octopusV1alpha1.ServicePort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     port.Port,
		})
	}
	return ports
}

func (ctr *Controller) getObjectNameWithClusterID(name, namespace string) string {
	return name + "-" + namespace + "-" + ctr.clusterID
}

func (ctr *Controller) serviceSyncLister(tranform func(ss *octopusV1alpha1.ServiceSync) runtime.Object) []runtime.Object {
	ssList, err := ctr.serviceSyncSyncer.ListLocalResources(&octopusV1alpha1.ServiceSync{})
	if err != nil {
		klog.Errorf("Error listing servicesync: %v", err)
	}

	retList := make([]runtime.Object, 0, len(ssList))
	for _, obj := range ssList {
		ss := obj.(*octopusV1alpha1.ServiceSync)

		retList = append(retList, tranform(ss))
	}
	return retList
}
