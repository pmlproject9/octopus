package multinetworkpolicy

import (
	"context"
	"github.com/pmlproject9/octopus/pkg/ipset"
	"github.com/pmlproject9/octopus/pkg/iptables"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	utilexec "k8s.io/utils/exec"
	"sync"
	"time"

	octopusapi "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	octopusClient "github.com/pmlproject9/octopus/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	ctx           context.Context
	kubeClient    kubernetes.Interface
	octopusClient octopusClient.Interface
	syncPeriod    time.Duration
	cluterID      string

	fullSyncChan chan struct{}
	reFlushChan  chan struct{}

	iptablesRunner *iptables.Executor
	ipsetRunner    *ipset.Executor
	ipsetMutex     sync.Mutex

	namespaceLister          cache.Indexer
	podLister                cache.Indexer
	multiNetworkPolicyLister cache.Indexer
	serviceSyncLister        cache.Indexer
	endpointLister           cache.Indexer
}

func NewMultiNetworkPolicyController(ctx context.Context,
	syncPeriod time.Duration,
	kubeclient kubernetes.Interface,
	nsInformer cache.SharedIndexInformer,
	podInformer cache.SharedIndexInformer,
	mnpInformer cache.SharedIndexInformer,
	serviceSyncInformer cache.SharedIndexInformer,
	endpointsInformer cache.SharedIndexInformer,
	clusterID string) *Controller {

	nsLister := nsInformer.GetIndexer()
	mnpLister := mnpInformer.GetIndexer()
	podLister := podInformer.GetIndexer()
	serviceSyncLister := serviceSyncInformer.GetIndexer()
	endpointsLister := endpointsInformer.GetIndexer()

	iptablesRunner, err := iptables.New()
	if err != nil {
		klog.Fatalf("error to create iptables executor:%v", err)
	}
	ipsetRunner := ipset.NewExecutor(utilexec.New())

	c := &Controller{
		ctx:        ctx,
		kubeClient: kubeclient,
		syncPeriod: syncPeriod,
		cluterID:   clusterID,

		fullSyncChan: make(chan struct{}, 1),
		reFlushChan:  make(chan struct{}, 1),

		iptablesRunner: iptablesRunner,
		ipsetRunner:    ipsetRunner,

		namespaceLister:          nsLister,
		multiNetworkPolicyLister: mnpLister,
		podLister:                podLister,
		serviceSyncLister:        serviceSyncLister,
		endpointLister:           endpointsLister,
	}

	mnpInformer.AddEventHandler(c.newMultiNetworkPolicyEventHandler())
	nsInformer.AddEventHandler(c.newNamespaceEventHandler())
	podInformer.AddEventHandler(c.newPodventHandler())
	serviceSyncInformer.AddEventHandler(c.newServiceSyncvEventHandler())
	endpointsInformer.AddEventHandler(c.newEndpointventHandler())

	return c
}

func (ctr *Controller) Run(wg *sync.WaitGroup) {
	t := time.NewTicker(ctr.syncPeriod)

	ctr.ensureDefaultIPtables()

	wg.Add(2)
	go ctr.FullSync(wg)
	go ctr.ReFlushServicesCidr(wg)

	for {
		klog.V(1).Info()
		ctr.fullSync()
		select {
		case <-ctr.ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (ctr *Controller) SendFullSyncChan() {
	select {
	case ctr.fullSyncChan <- struct{}{}:
		klog.V(4).Infof("send a message to fullsync to update multinetworkpolicy")
	default:
		klog.V(1).Infof("fullsyncchan is full , so skip")
	}

}

func (ctr *Controller) SendReFlushChan() {
	select {
	case ctr.reFlushChan <- struct{}{}:
		klog.V(4).Infof("send a message to reflush to update ipset")
	default:
		klog.V(1).Infof("reflushChan is full, so skip")
	}
}

func (ctr *Controller) newMultiNetworkPolicyEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: ctr.OnMultiNetworkPolicyCreate,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctr.OnMultiNetworkPolicyCreate(newObj)
		},
		DeleteFunc: ctr.OnMultiNetworkPolicyDelete,
	}
}

func (ctr *Controller) OnMultiNetworkPolicyCreate(obj interface{}) {
	mnp := obj.(*octopusapi.MultiNetworkPolicy)
	klog.V(4).Infof("adding MultiNetworkPolicyCreateRequest %q", klog.KObj(mnp))
	if err := ctr.ensureMultiNetworkPolicy(mnp); err != nil {
		klog.Errorf("error to update multinetworkpolicy %s: %v", mnp.Namespace+"/"+mnp.Name, err)
	}
}

func (ctr *Controller) OnMultiNetworkPolicyDelete(obj interface{}) {
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
	klog.V(4).Infof("deleting multiNetworkPolicy %q", klog.KObj(mnp))
	err := ctr.deleteMultiNetworkPolicy(mnp)
	if err != nil {
		klog.Errorf("error to delete multiNetworkPolicy %a", klog.KObj(mnp))
	}
}
