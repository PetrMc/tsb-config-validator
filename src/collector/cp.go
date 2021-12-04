package collector

import (
	"context"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type CPInterface interface {
	Get(name string) (*v1alpha1.ControlPlane, error)
}

type CPClient struct {
	restClient rest.Interface
	ns         string
}

func (c *CPClient) Get(name string) (*v1alpha1.ControlPlane, error) {

	result := v1alpha1.ControlPlane{}
	ctx := context.Background()
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("controlplanes").
		Name(name).
		Do(ctx).
		Into(&result)

	return &result, err
}
