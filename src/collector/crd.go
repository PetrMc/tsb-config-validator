package collector

import (
	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"
)

// Based on rest.config the function creates a client to 
// query CP CRD
func NewForConfig(c *rest.Config) (*K8sClient, error) {

	// adding API definition
	controlplane.AddToScheme(scheme.Scheme)

	crdConfig := *c
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: controlplane.GroupName, Version: controlplane.GroupVersion}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.UnversionedRESTClientFor(&crdConfig)

	if err != nil {
		return nil, err
	}

	return &K8sClient{restClient: client}, nil
}

// CP joins rest.Client config and namespace definitions
func (c *K8sClient) CP(namespace string) CPInterface {
	return &CPClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
