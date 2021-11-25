// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

package v1alpha1

import (
	pb "github.com/PetrMc/tsb-config-validator/api/install/controlplane/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlane is the Schema for the controlplanes API
// +kubebuilder:resource:path=controlplanes,scope=Namespaced,shortName=control,singular=controlplane
// +kubebuilder:object:root=true
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *pb.ControlPlaneSpec `json:"spec,omitempty"`
}

// ControlPlaneList contains a list of ControlPlane
// +kubebuilder:object:root=true
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
