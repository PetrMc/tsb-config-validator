package collector

import (
	"fmt"
	"os"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
	"k8s.io/client-go/kubernetes"
)

// var ElasticSearchSettings_Protocol_name = map[int32]string{
// 	0: "https",
// 	1: "http",
// }

type CPTelemetryStore struct {
	Host, Port, Protocol, Version string
	SelfSigned                    bool
}

type TSBConf struct {
	Host, Port, ClusterName string
}

func CPData() (ES, CPTelemetryStore, TSBConf, TSBTokens) {

	// config is used to get current context

	config = k8s()

	// cpsc is used from k8s `config` for CP access
	cpcs, err := NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(123)
	}

	// get details of CP CRD variables
	cp, err := cpcs.CP("istio-system").Get("controlplane")
	if err != nil {

		fmt.Println(err)
		os.Exit(123)
	}

	// there is slightly different config that is required to collect secrets
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Println(err)
		os.Exit(123)
	}

	creds := Secrets(clientset, "istio-system")
	tokens := TokensAll(clientset)

	var conn CPTelemetryStore
	conn, tsb := parameters(cp)

	return creds, conn, tsb, tokens

}

func parameters(c *controlplane.ControlPlane) (CPTelemetryStore, TSBConf) {
	var j CPTelemetryStore = telemetry(c)
	var k TSBConf = tsb(c)

	return j, k
}

func telemetry(c *controlplane.ControlPlane) CPTelemetryStore {
	b := CPTelemetryStore{}

	b.Host = c.Spec.TM.Elastic.Host
	b.Port = fmt.Sprint(c.Spec.TM.Elastic.Port)
	// b.Protocol = ElasticSearchSettings_Protocol_name[c.Spec.TM.Elastic.Protocol]
	if c.Spec.TM.Elastic.Protocol == "" {
		b.Protocol = "https"
	} else {
		b.Protocol = c.Spec.TM.Elastic.Protocol
	}
	b.SelfSigned = c.Spec.TM.Elastic.SelfSigned
	b.Version = fmt.Sprint(c.Spec.TM.Elastic.Version)

	return b
}

func tsb(c *controlplane.ControlPlane) TSBConf {
	b := TSBConf{}

	b.Host = c.Spec.MP.Host
	b.Port = fmt.Sprint(c.Spec.MP.Port)
	b.ClusterName = c.Spec.MP.ClusterName

	return b
}
