package v1alpha1

import (
	"k8s.io/client-go/rest"
)

type CPInterface interface {
	// List(opts metav1.ListOptions) (*cp.ControlPlaneList, error)
	// Get(name string, options metav1.GetOptions) (*cp.ControlPlane, error)
	// Create(*cp.ControlPlane) (*cp.ControlPlane, error)
	// Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type CPClient struct {
	restClient rest.Interface
	ns         string
}

// func (c *CPClient) List(opts metav1.ListOptions) (*cp.ControlPlaneList, error) {
// 	result := cp.ControlPlaneList{}
// 	err := c.restClient.
// 		Get().
// 		Namespace(c.ns).
// 		Resource("controlplane").
// 		VersionedParams(&opts, scheme.ParameterCodec).
// 		Do().
// 		Into(&result)

// 	return &result, err
// }

// func (c *CPClient) Get(name string, opts metav1.GetOptions) (*cp.ControlPlane, error) {
// 	result := cp.ControlPlane{}
// 	err := c.restClient.
// 		Get().
// 		Namespace(c.ns).
// 		Resource("controlplane").
// 		Name(name).
// 		VersionedParams(&opts, scheme.ParameterCodec).
// 		Do().
// 		Into(&result)

// 	return &result, err
// }

// func (c *CPClient) Create(project *cp.ControlPlane) (*cp.ControlPlane, error) {
// 	result := cp.ControlPlane{}
// 	err := c.restClient.
// 		Post().
// 		Namespace(c.ns).
// 		Resource("controlplane").
// 		Body(project).
// 		Do().
// 		Into(&result)

// 	return &result, err
// }

// func (c *CPClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
// 	opts.Watch = true
// 	return c.restClient.
// 		Get().
// 		Namespace(c.ns).
// 		Resource("controlplane").
// 		VersionedParams(&opts, scheme.ParameterCodec).
// 		Watch()
// }
