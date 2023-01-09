package multinetworkpolicy

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	octopusv1alpha1 "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	"github.com/pmlproject9/octopus/pkg/constants"
	listersoctopusv1alpha1 "github.com/pmlproject9/octopus/pkg/generated/listers/octopus.io/v1alpha1"
	"github.com/pmlproject9/octopus/pkg/ipset"
	submv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	utilexec "k8s.io/utils/exec"
)

func (ctr *Controller) FullSync(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctr.ctx.Done():
			klog.V(4).Info("shutting down the fullsync grouting")
			return
		default:
		}

		select {
		case <-ctr.fullSyncChan:
			ctr.fullSync()
		case <-ctr.ctx.Done():
			klog.V(4).Info("shutting down the fullsync grouting")
			return

		}
	}
}

func (ctr *Controller) fullSync() {
	ctr.ensureDefaultIPtables()

	ctr.ensureAllMultiNetworkPolicy()

	ctr.cleanUpStaleIptablesAndIPSet()
}

func (ctr *Controller) ReFlushServicesCidr(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctr.ctx.Done():
			klog.V(4).Info("shutting down the fullsync grouting")
			return
		default:
		}

		select {
		case <-ctr.reFlushChan:
			err := ctr.reFlushServiceCidrIPSet()
			if err != nil {
				klog.Errorf("error to update ipset %s : %v", constants.IPSetServicePrefix, err)
			}
		case <-ctr.ctx.Done():
			klog.V(4).Info("shutting down the reflush ipset grouting")
			return
		}
	}
}

func (ctr *Controller) reFlushServiceCidrIPSet() error {
	ips := ipset.New(constants.IPSetAllServicesCidr, ipset.HashNet, nil)
	ipSetExecutor := ipset.NewExecutor(utilexec.New())
	err := ipSetExecutor.Create(ips)
	if err != nil {
		return err
	}

	entries := make([]string, 0)
	eps := ctr.endpointLister.List()

	for _, epObj := range eps {
		ep, ok := epObj.(*submv1.Endpoint)
		if !ok {
			klog.Errorf("error to convert endpoint")
		}
		if ep.Spec.ClusterID == ctr.cluterID {
			continue
		}
		entries = append(entries, ep.Spec.Subnets...)
	}

	err = ipSetExecutor.ReFlush(ips, entries)
	return err
}

func (ctr *Controller) ensureDefaultIPtables() {
	err := ctr.iptablesRunner.CreateChainIfNotExist(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFORWARD)
	if err != nil {
		klog.Errorf("error to create iptables chain %s", constants.IPTablesChainOCTOPUSFORWARD)
	}
	err = ctr.iptablesRunner.CreateChainIfNotExist(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD)
	if err != nil {
		klog.Errorf("error to create iptables chain %s", constants.IPTablesChainOCTOPUSMNPFORWARD)
	}
	forwardToOctopusForward := []string{"-m", "comment", "--comment", "OCTOPUS", "-j", "OCTOPUS-FORWARD"}
	if err = ctr.iptablesRunner.InsertUnique(constants.IPTablesFilter, "FORWARD", 1, forwardToOctopusForward); err != nil {
		klog.Errorf("error to insert iptables chain %s rule %s : %v", "FORWARD", strings.Join(forwardToOctopusForward, " "), err)
	}
	forwardToMNPForward := []string{"-m", "set", "--match-set", constants.IPSetAllServicesCidr, "dst", "-j", constants.IPTablesChainOCTOPUSMNPFORWARD}
	if err = ctr.iptablesRunner.InsertUnique(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFORWARD, 1, forwardToMNPForward); err != nil {
		klog.Errorf("error to insert iptables chain %s rule %s : %v", "FORWARD", strings.Join(forwardToMNPForward, " "), err)
	}
	MNPStart := []string{"-m", "comment", "--comment", "start of multicluster policies", "-j", "MARK", "--set-xmark", "0x0/0x200000"}
	if err = ctr.iptablesRunner.InsertUnique(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD, 1, MNPStart); err != nil {
		klog.Errorf("error to insert iptables chain %s rule %s: %v", constants.IPTablesChainOCTOPUSMNPFORWARD, strings.Join(MNPStart, " "), err)
	}

	MNPEnd := []string{"-m", "comment", "--comment", "Drop if no multi network policies passed packed", "-m", "mark", "--mark", "0x0/0x200000", "-j", "DROP"}
	if err = ctr.iptablesRunner.InsertAtEnd(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD, MNPEnd); err != nil {
		klog.Errorf("error to insert iptables chain %s rule %s: %v", constants.IPTablesChainOCTOPUSMNPFORWARD, strings.Join(MNPEnd, " "), err)
	}
}

func (ctr *Controller) cleanUpStaleIptablesAndIPSet() {
	ctr.ipsetMutex.Lock()
	defer ctr.ipsetMutex.Unlock()
	activeIPset := make(map[string]bool)
	activeIPtablesChain := make(map[string]bool)
	for _, mnpObj := range ctr.multiNetworkPolicyLister.List() {
		mnp, ok := mnpObj.(*octopusv1alpha1.MultiNetworkPolicy)
		if !ok {
			klog.Errorf("error to convert multinetworkpolicy")
		}
		mnpNameHash := getMultiNetworkPolicyNameHash(mnp.Namespace, mnp.Name)
		activeIPset[constants.IPSetPodPrefix+mnpNameHash] = true
		activeIPset[constants.IPSetServicePrefix+mnpNameHash] = true
		activeIPtablesChain[constants.IPTablesChainOCTOPUSFWPREFIX+mnpNameHash] = true
	}
	ipsetList, err := ctr.ipsetRunner.ListIPSets()
	if err != nil {
		klog.Errorf("error to list ipset: %v", err)
	}
	for _, ips := range ipsetList {
		if strings.Contains(ips, constants.IPSetPodPrefix) || strings.Contains(ips, constants.IPSetServicePrefix) {
			if _, ok := activeIPset[ips]; !ok {
				ctr.ipsetRunner.Destroy(ips)
			}
		}
	}
	iptablesChains, err := ctr.iptablesRunner.ListChains(constants.IPTablesFilter)
	if err != nil {
		klog.Errorf("error list iptables chain from filter")
	}
	iptablesRules, err := ctr.iptablesRunner.List(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD)
	if err != nil {
		klog.Errorf("error list iptables rules from chain %s", constants.IPTablesChainOCTOPUSFORWARD)
	}
	for _, chain := range iptablesChains {
		if strings.Contains(chain, constants.IPTablesChainOCTOPUSFWPREFIX) {
			if _, ok := activeIPtablesChain[chain]; !ok {
				if err := ctr.iptablesRunner.DeleteChainDirect(constants.IPTablesFilter, chain); err != nil {
					klog.Errorf("error to delete chain %s:%v", chain, err)
				}
			}
		}
	}

	for _, rule := range iptablesRules {
		if !strings.Contains(rule, "-m set --match-set") {
			continue
		}
		ruleSplit := strings.Split(rule, " ")
		targetChain := ruleSplit[len(ruleSplit)-1]
		if _, ok := activeIPtablesChain[targetChain]; !ok {
			if err := ctr.iptablesRunner.Delete(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD, ruleSplit[2:]...); err != nil {
				klog.Errorf("error to delete rule %s", strings.Join(ruleSplit[2:], " "))
			}
		}
	}
}

func getMultiNetworkPolicyNameHash(namespace string, name string) string {
	hash := sha256.Sum256([]byte(namespace + name))
	encode := base32.StdEncoding.EncodeToString(hash[:])
	return encode[:16]
}

func (ctr *Controller) ensureAllMultiNetworkPolicy() {

	for _, mnpObj := range ctr.multiNetworkPolicyLister.List() {
		mnp, ok := mnpObj.(*octopusv1alpha1.MultiNetworkPolicy)
		if !ok {
			klog.Errorf("error to convert multinetworkpolicy\n")
		}
		if err := ctr.ensureMultiNetworkPolicy(mnp); err != nil {
			klog.Errorf("%v", err)
		}
	}
}

func (ctr *Controller) ensureMultiNetworkPolicy(mnp *octopusv1alpha1.MultiNetworkPolicy) error {
	err := ctr.ensureMultiNetworkPolicyIPsets(mnp)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error to convert multinetworkpolicy %s into ipset: %v", mnp.Namespace+"/"+mnp.Name, err))
	}
	err = ctr.ensureMultiNetworkPolicyIPtables(mnp)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error to convert multinetworkpolicy %s into iptable rule:%v", mnp.Namespace+"/"+mnp.Name, err))
	}
	return nil
}

func (ctr *Controller) deleteMultiNetworkPolicy(mnp *octopusv1alpha1.MultiNetworkPolicy) error {
	ctr.ipsetMutex.Lock()
	defer ctr.ipsetMutex.Unlock()

	mnpNameHash := getMultiNetworkPolicyNameHash(mnp.Namespace, mnp.Name)
	ipsetPodName := constants.IPSetPodPrefix + mnpNameHash
	ipsetServiceName := constants.IPSetServicePrefix + mnpNameHash
	iptableName := constants.IPTablesChainOCTOPUSFWPREFIX + mnpNameHash

	args := []string{"-m", "set", "--match-set", ipsetPodName, "src", "-j", iptableName}
	if err := ctr.iptablesRunner.Delete(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD, args...); err != nil {
		return errors.Wrap(err, "error to delete iptable rule")
	}
	if err := ctr.iptablesRunner.DeleteChainDirect(constants.IPTablesFilter, iptableName); err != nil {
		return errors.Wrap(err, "error to delete chain "+iptableName)
	}
	if err := ctr.ipsetRunner.Destroy(ipsetPodName); err != nil {
		return errors.Wrap(err, "error delete ipset "+ipsetPodName)
	}
	if err := ctr.ipsetRunner.Destroy(ipsetServiceName); err != nil {
		return errors.Wrap(err, "error to delete ipset "+ipsetServiceName)
	}
	return nil
}

func (ctr *Controller) ensureMultiNetworkPolicyIPsets(mnp *octopusv1alpha1.MultiNetworkPolicy) error {
	ctr.ipsetMutex.Lock()
	defer ctr.ipsetMutex.Unlock()

	mnpHashName := getMultiNetworkPolicyNameHash(mnp.Namespace, mnp.Name)
	podSelector, _ := metav1.LabelSelectorAsSelector(&mnp.Spec.PodSelector)
	podIPsetName := constants.IPSetPodPrefix + mnpHashName
	serviceIPsetName := constants.IPSetServicePrefix + mnpHashName

	pods, _ := ctr.ListPodsByNamesapceAndLabels(mnp.Namespace, podSelector)
	podEntries := make([]string, 0)
	for _, pod := range pods {
		if isNetPolActionable(pod) {

		}
		podEntries = append(podEntries, pod.Status.PodIP)
	}
	if err := ctr.ipsetRunner.ReFlush(ipset.New(podIPsetName, ipset.HashNet, nil), podEntries); err != nil {
		klog.Errorf("error to reflush ipset %s: %v", podIPsetName, err)
	}
	servicesEntriesMap := make(map[string]bool)
	for _, selectors := range mnp.Spec.Egress.Allow {
		if selectors.NamespaceSelector != nil {
			nsSelector, _ := metav1.LabelSelectorAsSelector(selectors.NamespaceSelector)
			nss, _ := ctr.ListNamespaceByLabels(nsSelector)
			for _, ns := range nss {
				services, _ := ctr.ListServiceSyncByLabels(ns.Name, labels.Everything())
				for _, service := range services {
					if service.Labels[constants.LabelSourceClusterID] == ctr.cluterID {
						continue
					}
					for _, ip := range service.Spec.IPs {
						servicesEntriesMap[ip] = true
					}
				}
			}
		}
		if selectors.ServiceSelector != nil {
			serviceSelector, _ := metav1.LabelSelectorAsSelector(selectors.ServiceSelector)
			services, _ := ctr.ListServiceSyncByLabels(mnp.Namespace, serviceSelector)
			for _, service := range services {
				if service.Labels[constants.LabelSourceClusterID] == ctr.cluterID {
					continue
				}
				for _, ip := range service.Spec.IPs {
					servicesEntriesMap[ip] = true
				}
			}
		}
	}
	servicesEntries := make([]string, 0)
	for k := range servicesEntriesMap {
		servicesEntries = append(servicesEntries, k)
	}
	if err := ctr.ipsetRunner.ReFlush(ipset.New(serviceIPsetName, ipset.HashNet, nil), servicesEntries); err != nil {
		klog.Errorf("error to reflush ipset %s: %v", serviceIPsetName, err)
	}
	return nil
}

func (ctr *Controller) ensureMultiNetworkPolicyIPtables(mnp *octopusv1alpha1.MultiNetworkPolicy) error {
	mnpHashName := getMultiNetworkPolicyNameHash(mnp.Namespace, mnp.Name)
	if err := ctr.iptablesRunner.CreateChainIfNotExist(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName); err != nil {
		klog.Errorf("error to create chain %s", constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName)
	}
	args := []string{"-m", "set", "--match-set", constants.IPSetPodPrefix + mnpHashName, "src", "-j", constants.IPTablesChainOCTOPUSFWPREFIX + mnpHashName}
	err := ctr.iptablesRunner.AppendUnique(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSMNPFORWARD, args...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error to append rule info chain %s", constants.IPTablesChainOCTOPUSMNPFORWARD))
	}
	err = ctr.iptablesRunner.ClearChain(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error to clear rule from chain %s", constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName))
	}
	args = []string{"-m", "set", "--match-set", constants.IPSetServicePrefix + mnpHashName, "dst"}
	argsEnd := []string{"-j", "ACCEPT"}
	if len(mnp.Spec.Egress.Ports) > 0 {
		ports := mnp.Spec.Egress.Ports

		for _, port := range ports {
			argsCopy := []string{}
			argsCopy = append(argsCopy, args...)
			if port.Protocol != nil {
				argsCopy = append(argsCopy, "-p", string(*port.Protocol))
			}
			argsCopy = append(argsCopy, "-m", "multiport")
			if port.EndPort != nil {
				argsCopy = append(argsCopy, "--dport", port.Port.String()+":"+strconv.Itoa(int(*port.EndPort)))
			} else {
				argsCopy = append(argsCopy, "--dport", port.Port.String())
			}
			argsCopy = append(argsCopy, argsEnd...)
			err = ctr.iptablesRunner.Append(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName, argsCopy...)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error to append rule into %s", constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName))
			}
		}

	} else {
		args = append(args, argsEnd...)
		err = ctr.iptablesRunner.Append(constants.IPTablesFilter, constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName, args...)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error to append rule into %s", constants.IPTablesChainOCTOPUSFWPREFIX+mnpHashName))
		}
	}
	return nil
}

func (ctr *Controller) ListPodsByNamesapceAndLabels(namespace string, label labels.Selector) (ret []*corev1.Pod, err error) {
	podLister := listerscorev1.NewPodLister(ctr.podLister)
	pod, err := podLister.Pods(namespace).List(label)
	if err != nil {
		return nil, errors.Wrap(err, "error to list pod")
	}
	return pod, nil
}

func (ctr *Controller) ListNamespaceByLabels(label labels.Selector) (ret []*corev1.Namespace, err error) {
	namespaceLister := listerscorev1.NewNamespaceLister(ctr.namespaceLister)
	ns, err := namespaceLister.List(label)
	if err != nil {
		return nil, errors.Wrap(err, "error to list namespace")
	}
	return ns, nil
}

func (ctr *Controller) ListServiceSyncByLabels(namespace string, label labels.Selector) (ret []*octopusv1alpha1.ServiceSync, err error) {
	serviceSyncLister := listersoctopusv1alpha1.NewServiceSyncLister(ctr.serviceSyncLister)
	serviceSyncs, err := serviceSyncLister.ServiceSyncs(namespace).List(label)
	if err != nil {
		return nil, errors.Wrap(err, "error to list servicesync")
	}
	return serviceSyncs, err
}

func isNetPolActionable(pod *corev1.Pod) bool {
	return !isFinished(pod) && pod.Status.PodIP != "" && !pod.Spec.HostNetwork
}

func isFinished(pod *corev1.Pod) bool {
	//nolint:exhaustive // We don't care about PodPending, PodRunning, PodUnknown here as we want those to fall
	// into the false case
	switch pod.Status.Phase {
	case corev1.PodFailed, corev1.PodSucceeded, "Completed":
		return true
	}
	return false
}
