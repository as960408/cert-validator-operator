/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertValidateSpec defines the desired state of CertValidate.
type CertValidateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CertValidate. Edit certvalidate_types.go to remove/update
	NodeName string `json:"nodeName"`
	FilePath string `json:"filePath"`
	Expiry   string `json:"expiry"` // ISO8601 형식: 2025-06-01T12:00:00Z
	Valid    bool   `json:"valid"`
}

// CertValidateStatus defines the observed state of CertValidate.
type CertValidateStatus struct {
	Message string `json:"message,omitempty"` // optional

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="expirity",type=string,JSONPath=`.spec.expiry`
// +kubebuilder:printcolumn:name="valid",type=string,JSONPath=`.spec.valid`
// +kubebuilder:printcolumn:name="age",type=date,JSONPath=`.metadata.creationTimestamp`
// CertValidate is the Schema for the certvalidates API.
type CertValidate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertValidateSpec   `json:"spec,omitempty"`
	Status CertValidateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertValidateList contains a list of CertValidate.
type CertValidateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertValidate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertValidate{}, &CertValidateList{})
}
