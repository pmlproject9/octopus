package multinetworkpolicy

import (
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"

	submv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
)

func (ctr *Controller) newEndpointventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.OnEndpointCreate,
		UpdateFunc: ctr.OnEndpointUpdate,
		DeleteFunc: ctr.OnEndpointDelete,
	}
}

func (ctr *Controller) OnEndpointCreate(obj interface{}) {
	ep, ok := obj.(*submv1.Endpoint)
	if !ok {
		return
	}
	klog.V(4).Infof("adding endpointCreateRequest %q", klog.KObj(ep))
	ctr.SendReFlushChan()
}

func (ctr *Controller) OnEndpointUpdate(oldObj interface{}, newObj interface{}) {
	oldEp, ok := oldObj.(*submv1.Endpoint)
	if !ok {
		return
	}
	newEp, ok := newObj.(*submv1.Endpoint)
	if !ok {
		return
	}

	if reflect.DeepEqual(newEp.Spec.Subnets, oldEp.Spec.Subnets) {
		klog.V(4).Infof("no labels update in pod,name is %s, skipping sync", newEp.Name)
		return
	}
	klog.V(4).Infof("updating endpointUpdateRequest %q", klog.KObj(newEp))
	ctr.SendReFlushChan()
}

func (ctr *Controller) OnEndpointDelete(obj interface{}) {
	ep, ok := obj.(*submv1.Endpoint)
	if !ok {
		tombstone, ok := obj.(*cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("couldn't get object from tombstone: %v", obj)
			return
		}
		ep, ok = tombstone.Obj.(*submv1.Endpoint)
		if !ok {
			klog.Errorf("tombstone contained object that is not pod: %v", obj)
			return
		}
	}
	klog.V(4).Infof("delete endpoint: %q", klog.KObj(ep))
	ctr.SendReFlushChan()
}
