package multinetworkpolicy

import (
	octopusapi "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ctr *Controller) newNamespaceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.OnNamespaceCreate,
		UpdateFunc: ctr.OnNamespaceUpdate,
		DeleteFunc: ctr.OnNamespeceDelete,
	}
}

func (ctr *Controller) OnNamespaceCreate(obj interface{}) {
	ns := obj.(*corev1.Namespace)
	klog.V(4).Infof("adding namespaceCreateRequest %q", klog.KObj(ns))
	if ns.Labels == nil {
		klog.V(4).Infof("namespace has no labels, skipping sync")
	}
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnNamespaceUpdate(oldObj interface{}, newObj interface{}) {
	oldNs := oldObj.(*corev1.Namespace)
	newNs := newObj.(*corev1.Namespace)

	if reflect.DeepEqual(oldNs.Labels, newNs.Labels) {
		klog.V(4).Infof("no labels update in namespace,name is %s, skipping sync", newNs.Name)
		return
	}
	klog.V(4).Infof("updating namespaceCreateRequest %q", klog.KObj(newNs))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnNamespeceDelete(obj interface{}) {
	mnp, ok := obj.(*octopusapi.MultiNetworkPolicy)
	if !ok {
		tombstone, ok := obj.(*cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("couldn't get object from tombstone: %v", obj)
			return
		}
		mnp, ok = tombstone.Obj.(*octopusapi.MultiNetworkPolicy)
		if !ok {
			klog.Errorf("tombstone contained object that is not MultiNetworkPolicy: %v", obj)
			return
		}
	}
	klog.V(4).Infof("", klog.KObj(mnp))
	ctr.SendFullSyncChan()

}
