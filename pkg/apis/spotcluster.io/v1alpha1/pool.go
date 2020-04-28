package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=pool
// +k8s:openapi-gen=true
// +genclient:nonNamespaced

type Pool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec         ClusterSpec  `json:"spec,omitempty"`
	ProviderSpec ProviderSpec `json:"providerSpec,omitempty"`
}

type ClusterSpec struct {
	Replicas       int    `json:"replicas,omitempty"`
	SSHFingerprint string `json:"sshFingerprint,omitempty"`
	NodeToken      string `json:"nodeToken,omitempty"`
	MasterURL      string `json:"masterUrl,omitempty"`
}

type ProviderSpec struct {
	DigitalOcean *DigitalOcean `json:"digitalOcean,omitempty"`
}

type DigitalOcean struct {
	Image        string `json:"image,omitempty"`
	InstanceSize string `json:"instanceSize,omitempty"`
	APIKey       string `json:"apiKey,omitempty"`
	Region       string `json:"region,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=poolList
// +k8s:openapi-gen=true

type PoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pool `json:"items"`
}
