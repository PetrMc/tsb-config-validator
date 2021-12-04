package collector

import (
	"github.com/PetrMc/tsb-config-validator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type CPTelemetryStore struct {
	Host, Port, Version string
	SSL                 bool
}

type K8sInterface interface {
	CP(namespace string) CPInterface
}

type K8sClient struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*K8sClient, error) {
	v1alpha1.AddToScheme(scheme.Scheme)
	crdconfig := *config
	crdconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.SchemeGroupVersion.Group, Version: v1alpha1.SchemeGroupVersion.Version}
	crdconfig.APIPath = "/apis"
	crdconfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdconfig.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.UnversionedRESTClientFor(&crdconfig)

	if err != nil {
		return nil, err
	}

	return &K8sClient{restClient: client}, nil
}

func (c *K8sClient) CP(namespace string) CPInterface {
	return &CPClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
