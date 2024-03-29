/*
Copyright The octopus Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/pmlproject9/octopus/pkg/apis/octopus.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ServiceSyncLister helps list ServiceSyncs.
// All objects returned here must be treated as read-only.
type ServiceSyncLister interface {
	// List lists all ServiceSyncs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ServiceSync, err error)
	// ServiceSyncs returns an object that can list and get ServiceSyncs.
	ServiceSyncs(namespace string) ServiceSyncNamespaceLister
	ServiceSyncListerExpansion
}

// serviceSyncLister implements the ServiceSyncLister interface.
type serviceSyncLister struct {
	indexer cache.Indexer
}

// NewServiceSyncLister returns a new ServiceSyncLister.
func NewServiceSyncLister(indexer cache.Indexer) ServiceSyncLister {
	return &serviceSyncLister{indexer: indexer}
}

// List lists all ServiceSyncs in the indexer.
func (s *serviceSyncLister) List(selector labels.Selector) (ret []*v1alpha1.ServiceSync, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ServiceSync))
	})
	return ret, err
}

// ServiceSyncs returns an object that can list and get ServiceSyncs.
func (s *serviceSyncLister) ServiceSyncs(namespace string) ServiceSyncNamespaceLister {
	return serviceSyncNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceSyncNamespaceLister helps list and get ServiceSyncs.
// All objects returned here must be treated as read-only.
type ServiceSyncNamespaceLister interface {
	// List lists all ServiceSyncs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ServiceSync, err error)
	// Get retrieves the ServiceSync from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ServiceSync, error)
	ServiceSyncNamespaceListerExpansion
}

// serviceSyncNamespaceLister implements the ServiceSyncNamespaceLister
// interface.
type serviceSyncNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceSyncs in the indexer for a given namespace.
func (s serviceSyncNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ServiceSync, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ServiceSync))
	})
	return ret, err
}

// Get retrieves the ServiceSync from the indexer for a given namespace and name.
func (s serviceSyncNamespaceLister) Get(name string) (*v1alpha1.ServiceSync, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("servicesync"), name)
	}
	return obj.(*v1alpha1.ServiceSync), nil
}
