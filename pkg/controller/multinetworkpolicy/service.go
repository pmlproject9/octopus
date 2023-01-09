package multinetworkpolicy

import (
	octopusV1alpha1 "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ctr *Controller) newServiceSyncvEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.OnPodCreate,
		UpdateFunc: ctr.OnPodUpdate,
		DeleteFunc: ctr.OnPodDelete,
	}
}

func (ctr *Controller) OnServiceSyncCreate(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	klog.V(4).Infof("adding namespaceCreateRequest %q", klog.KObj(pod))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnServiceSyncUpdate(oldObj interface{}, newObj interface{}) {
	oldServiceSync, ok := oldObj.(*octopusV1alpha1.ServiceSync)
	if !ok {
		return
	}
	newServiceSync, ok := newObj.(*octopusV1alpha1.ServiceSync)
	if !ok {
		return
	}

	if reflect.DeepEqual(oldServiceSync.Labels, newServiceSync.Labels) &&
		reflect.DeepEqual(oldServiceSync.Spec.IPs, newServiceSync.Spec.IPs) {
		klog.V(4).Infof("no labels update in servicesync,name is %s, skipping sync", newServiceSync.Name)
		return
	}
	klog.V(4).Infof("updating podUpdateRequest %q", klog.KObj(newServiceSync))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnServiecSyncDelete(obj interface{}) {
	serviceSync, ok := obj.(*octopusV1alpha1.ServiceSync)
	if !ok {
		tombstone, ok := obj.(*cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("couldn't get object from tombstone: %v", obj)
			return
		}
		serviceSync, ok = tombstone.Obj.(*octopusV1alpha1.ServiceSync)
		if !ok {
			klog.Errorf("tombstone contained object that is not serviceSync: %v", obj)
			return
		}
	}
	klog.V(4).Infof("", klog.KObj(serviceSync))
	ctr.SendFullSyncChan()
}
