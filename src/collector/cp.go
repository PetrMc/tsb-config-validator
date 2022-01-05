package collector

import (
	"context"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
)


// Get used restclient config to query k8s cluster and obtain CP CRD
func (c *CPClient) Get(name string) (*controlplane.ControlPlane, error) {

	// received data will be stored in result variable
	var result controlplane.ControlPlane

	// err is to store and return the error if arise
	var err error

	ctx := context.Background()
	err = c.restClient.
		Get().
		Namespace(c.ns).
		Resource("controlplanes").
		Name(name).
		Do(ctx).
		Into(&result)

	return &result, err
}
