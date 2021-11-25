package v1alpha1

import (
	cp "github.com/PetrMc/tsb-config-validator/apis/install/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type V1Alpha1Interface interface {
	ControlPlanes(namespace string) CPInterface
}

type V1Alpha1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*V1Alpha1Client, error) {

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: cp.SchemeGroupVersion.Group, Version: cp.SchemeGroupVersion.Version}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &V1Alpha1Client{restClient: client}, nil
}

func (c *V1Alpha1Client) ControlPlanes(namespace string) CPInterface {
	return &CPClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
