// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

package apis

import (
	topv1alpha1 "github.com/PetrMc/tsb-config-validator/apis/install/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, topv1alpha1.SchemeBuilder.AddToScheme)
}
