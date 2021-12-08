package collector

import (
	"context"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
	"k8s.io/client-go/rest"
)

type CPInterface interface {
	Get(name string) (*controlplane.ControlPlane, error)
}

type CPClient struct {
	restClient rest.Interface
	ns         string
}

func (c *CPClient) Get(name string) (*controlplane.ControlPlane, error) {

	result := controlplane.ControlPlane{}
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
