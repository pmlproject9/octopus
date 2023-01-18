package multinetworkpolicy

import (
	"github.com/pmlproject9/octopus/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	mcsv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

func (ctr *Controller) newServiceImportEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: ctr.OnServiceImportCreate,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctr.OnServiceImportCreate(newObj)
		},
		DeleteFunc: ctr.OnServiecImportDelete,
	}
}

func (ctr *Controller) OnServiceImportCreate(obj interface{}) {
	serviceImport, ok := obj.(*mcsv1a1.ServiceImport)
	if !ok {
		return
	}
	klog.V(4).Infof("adding serviceImportCreateRequest %q", klog.KObj(serviceImport))
	ctr.SendFullSyncChan()
	ctr.UpdateServiceImport(serviceImport)

}

func (ctr *Controller) OnServiecImportDelete(obj interface{}) {
	serviceImport, ok := obj.(*mcsv1a1.ServiceImport)
	if !ok {
		tombstone, ok := obj.(*cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("couldn't get object from tombstone: %v", obj)
			return
		}
		serviceImport, ok = tombstone.Obj.(*mcsv1a1.ServiceImport)
		if !ok {
			klog.Errorf("tombstone contained object that is not serviceImport: %v", obj)
			return
		}
	}
	klog.V(4).Infof("", klog.KObj(serviceImport))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) UpdateServiceImport(serviceImport *mcsv1a1.ServiceImport) {
	if serviceImport.Labels[constants.LabelSourceClusterID] != ctr.cluterID {
		return
	}
	namespace := serviceImport.Labels[constants.LabelSourceNamespace]
	name := serviceImport.Labels[constants.LabelSourceName]

	service, err := ctr.kubeClient.CoreV1().Services(namespace).Get(ctr.ctx, name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("error to get service %s/%s, %v", namespace, name, err)
	}
	newServiceImport := serviceImport.DeepCopy()

	serviceLabels := service.Labels
	for k, v := range serviceLabels {
		newServiceImport.Labels[k] = v
	}
	_, err = ctr.mcsClient.MulticlusterV1alpha1().ServiceImports(serviceImport.Namespace).Update(ctr.ctx, newServiceImport, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("error to update serviceImport %s/%s, %v", namespace, name, err)
	} else {
		klog.Infof("success to label serviceimport %s/%s", serviceImport.Namespace, serviceImport.Name)
	}
}
