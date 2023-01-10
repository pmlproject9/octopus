package multinetworkpolicy

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
)

func (ctr *Controller) newPodventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.OnPodCreate,
		UpdateFunc: ctr.OnPodUpdate,
		DeleteFunc: ctr.OnPodDelete,
	}
}

func (ctr *Controller) OnPodCreate(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	klog.V(4).Infof("adding podCreateRequest %q", klog.KObj(pod))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnPodUpdate(oldObj interface{}, newObj interface{}) {
	oldPod, ok := oldObj.(*corev1.Pod)
	if !ok {
		return
	}
	newPod, ok := newObj.(*corev1.Pod)
	if !ok {
		return
	}

	if reflect.DeepEqual(oldPod.Labels, newPod.Labels) {
		klog.V(4).Infof("no labels update in pod,name is %s, skipping sync", newPod.Name)
		return
	}
	klog.V(4).Infof("updating podUpdateRequest %q", klog.KObj(newPod))
	ctr.SendFullSyncChan()
}

func (ctr *Controller) OnPodDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		tombstone, ok := obj.(*cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("couldn't get object from tombstone: %v", obj)
			return
		}
		pod, ok = tombstone.Obj.(*corev1.Pod)
		if !ok {
			klog.Errorf("tombstone contained object that is not pod: %v", obj)
			return
		}
	}
	klog.V(4).Infof("Deleting podUpdateRequest %q", klog.KObj(pod))
	ctr.SendFullSyncChan()
}
