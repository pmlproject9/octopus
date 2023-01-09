package iptables

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"strings"
)

type Basic interface {
	Append(table, chain string, rulespec ...string) error
	AppendUnique(table, chain string, rulespec ...string) error
	Delete(table, chain string, rulespec ...string) error
	Insert(table, chain string, pos int, rulespec ...string) error
	List(table, chain string) ([]string, error)
	ListChains(table string) ([]string, error)
	NewChain(table, chain string) error
	ChainExists(table, chain string) (bool, error)
	ClearChain(table, chain string) error
	DeleteChain(table, chain string) error
}

type Executor struct {
	Basic
}

func New() (*Executor, error) {
	ipts, err := iptables.New()
	if err != nil {
		return nil, err
	}
	return &Executor{
		ipts,
	}, nil
}

func (e *Executor) CreateChainIfNotExist(table string, chian string) error {
	exist, err := e.ChainExists(table, chian)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	err = e.NewChain(table, chian)
	return err
}

func (e *Executor) InsertUnique(table string, chain string, pos int, ruleSpec []string) error {
	existRules, err := e.List(table, chain)
	if err != nil {
		klog.Errorf("error to list table %s chain %s", table, chain)
		return errors.Wrap(err, "error to list chain rule")
	}
	existCounts := 0
	posTrue := false
	ruleSpecString := strings.Join(ruleSpec, " ")
	for index, rule := range existRules {
		ruleSubs := strings.Split(rule, "\"")
		if strings.Contains(strings.Join(ruleSubs, ""), ruleSpecString) {
			existCounts++
		}
		if index == pos {
			posTrue = true
		}
	}
	if posTrue && existCounts == 1 {
		return nil
	}

	for i := 0; i < existCounts; i++ {
		err = e.Delete(table, chain, ruleSpec...)
		if err != nil {
			klog.Errorf("error to delete iptables rule %s from chain %s table %s: %v", ruleSpecString, chain, table, err)
			return errors.Wrap(err, "error to delete rule")
		}
	}
	err = e.Insert(table, chain, pos, ruleSpec...)
	return errors.Wrap(err, " error to insert rule")
}

func (e *Executor) InsertAtEnd(table string, chain string, ruleSpec []string) error {
	existRules, err := e.List(table, chain)
	if err != nil {
		klog.Errorf("error to list table %s chain %s", table, chain)
		return errors.Wrap(err, "error to list chain rule")
	}
	ruleSpecString := strings.Join(ruleSpec, " ")
	for i, rule := range existRules {
		ruleSubs := strings.Split(rule, "\"")
		if strings.Contains(strings.Join(ruleSubs, ""), ruleSpecString) {
			if i != len(existRules)-1 {
				if err := e.Delete(table, chain, ruleSpec...); err != nil {
					klog.Errorf("error to delete iptables rule %s from chain %s table %s: %v", ruleSpecString, chain, table, err)
				}
			} else {
				return nil
			}
		}
	}
	err = e.Append(table, chain, ruleSpec...)
	return errors.Wrap(err, "error to append rule")
}

func (e *Executor) DeleteChainDirect(table string, chain string) error {
	err := e.ClearChain(table, chain)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error to clear chain %s", chain))
	}
	err = e.DeleteChain(table, chain)
	return errors.Wrap(err, fmt.Sprintf("error to delete chain %s", chain))
}
