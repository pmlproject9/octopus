package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Important: Run "make generated" to regenerate code after modifying this file

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Namespaced",shortName=mnp
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true

// MultiNetworkPolicy is the Schema for the resources to be installed
type MultiNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec MultiNetworkPolicySpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

type MultiNetworkPolicySpec struct {
	PodSelector metav1.LabelSelector `json:"podSelector" protobuf:"bytes,1,opt, name=podSelector"`
	// +optional
	Egress Rule `json:"egress,omitempty" protobuf:"bytes,3,opt,name=egress"`
}

type Rule struct {
	Ports []Port `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`
	Allow []Peer `json:"allow,omitempty" protobuf:"bytes,2,rep,name=allow"`
}

type Port struct {
	// The protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
	// field defaults to TCP.
	// +optional
	Protocol *v1.Protocol `json:"protocol,omitempty" protobuf:"bytes,1,opt,name=protocol,casttype=k8s.io/api/core/v1.Protocol"`

	Port *intstr.IntOrString `json:"port,omitempty" protobuf:"bytes,2,opt,name=port"`

	// If set, indicates that the range of ports from port to endPort, inclusive,
	// should be allowed by the policy. This field cannot be defined if the port field
	// is not defined or if the port field is defined as a named (string) port.
	// The endPort must be equal or greater than port.
	// This feature is in Beta state and is enabled by default.
	// It can be disabled using the Feature Gate "NetworkPolicyEndPort".
	// +optional
	EndPort *int32 `json:"endPort,omitempty" protobuf:"bytes,3,opt,name=endPort"`
}

type Peer struct {
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,1,opt,name=namespaceSelector"`
	ServiceSelector   *metav1.LabelSelector `json:"serviceSelector,omitempty" protobuf:"bytes,1,opt,name=serviceSelector"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DescriptionList contains a list of Description
type MultiNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiNetworkPolicy `json:"items"`
}
