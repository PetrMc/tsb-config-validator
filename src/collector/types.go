package collector

import (
	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
	"k8s.io/client-go/rest"
)

// CPTelemetry store is utilized to store parameters
// that are declared in TSB Control Plane CRD "telemetry"
// section
type CPTelemetryStore struct {
	Host, Port, Protocol, Version string
	SelfSigned                    bool
}

// TSBConf struct stores "managementPlane section"
type TSBConf struct {
	Host, Port, ClusterName string
}



// CPInterface is defined for http queries to K8s
type CPInterface interface {
	Get(name string) (*controlplane.ControlPlane, error)
}

// CPClient joints rest.Client config and namespace of interest
type CPClient struct {
	restClient rest.Interface
	ns         string
}

// K8sInterface is created to handle interactions with K8s cluster
type K8sInterface interface {
	CP(namespace string) CPInterface
}

// K8sClient struct for restClient 
type K8sClient struct {
	restClient rest.Interface
}

// TSBTokens struct to organize the storage of tokens received from 
// CP CRD
type TSBTokens struct {
	Oapt, Otelt, Zipkint, Xcpt string
}

// ES stands for ElasticSearch - in this case keeping credentials and certificate together
// all used for http calls to ElasticSearch instance.
type ES struct {
	Username, Password, Cert string
}