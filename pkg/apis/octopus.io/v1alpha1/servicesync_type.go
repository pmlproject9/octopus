package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make generated" to regenerate code after modifying this file

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Namespaced",shortName=ss
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true

// ServiceSync is the Schema for the resources to be installed
type ServiceSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec ServiceSyncSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

type ServiceSyncSpec struct {
	IPs            []string           `json:"ips" protobuf:"bytes,1,rep, name=ips"`
	IPFamilies     []v1.IPFamily      `json:"ipFamilies,omitempty" protobuf:"bytes,2,rep, name=ipFamilies"`
	IPFamilyPolicy *v1.IPFamilyPolicy `json:"ipFamilyPolicy,omitempty" protobuf:"bytes,3,opt, name=ipFamilyPolicy"`

	// +optional
	Ports []ServicePort `json:"ports,omitempty" protobuf:"bytes,3,rep,name=ports"`
}

type ServicePort struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// The protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
	// field defaults to TCP.
	// +optional
	Protocol v1.Protocol `json:"protocol,omitempty" protobuf:"bytes,2,opt,name=protocol,casttype=k8s.io/api/core/v1.Protocol"`

	Port int32 `json:"port,omitempty" protobuf:"bytes,3,opt,name=port"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DescriptionList contains a list of Description
type ServiceSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceSync `json:"items"`
}
