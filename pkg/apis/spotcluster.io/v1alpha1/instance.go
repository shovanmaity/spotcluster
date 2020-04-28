package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=instance
// +k8s:openapi-gen=true
// +genclient:nonNamespaced

type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

type InstanceSpec struct {
	RemoteAddress     string `json:"remoteAddress,omitempty"`
	Provider          string `json:"provider,omitempty"`
	NodeName          string `json:"nodeName,omitempty"`
	InstanceName      string `json:"instanceName,omitempty"`
	InstanceAvailable bool   `json:"instanceAvailable,omitempty"`
	NodeAvailable     bool   `json:"nodeAvailable,omitempty"`
	InstanceReady     bool   `json:"instanceReady,omitempty"`
	NodeReady         bool   `json:"nodeReady,omitempty"`
}

type InstanceStatus struct {
	NodeStatus     string `json:"nodeStatus,omitempty"`
	InstanceStatus string `json:"instanceStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=instanceList
// +k8s:openapi-gen=true

type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}
