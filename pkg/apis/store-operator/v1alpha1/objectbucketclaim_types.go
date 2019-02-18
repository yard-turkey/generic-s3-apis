package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ObjectBucketClaimSpec defines the desired state of ObjectBucketClaim
type ObjectBucketClaimSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// ObjectBucketClaimStatus defines the observed state of ObjectBucketClaim
type ObjectBucketClaimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectBucketClaim is the Schema for the objectbucketclaims API
// +k8s:openapi-gen=true
type ObjectBucketClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObjectBucketClaimSpec   `json:"spec,omitempty"`
	Status ObjectBucketClaimStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectBucketClaimList contains a list of ObjectBucketClaim
type ObjectBucketClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectBucketClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObjectBucketClaim{}, &ObjectBucketClaimList{})
}
