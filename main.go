package main

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
	"github.com/PetrMc/tsb-config-validator/src/validator"
)

func main() {
	cred, conn, tsb := collector.CPData()

	printCP(&cred, &conn, &tsb)

	validator.Checklist(&cred, &conn)

	// fmt.Printf("Done main module.\n tsb port: %v\ncred username: %v pass: %v\n conn: %v\n", tsb, cred.Username, cred.Password, conn)

}

func printCP(c *collector.ES, n *collector.CPTelemetryStore, t *collector.TSBConf) {
	fmt.Printf("For Control plane: %v - the following data is received:\n", t.ClusterName)
	fmt.Printf("Management plane Front Envoy Address: %v Port: %v\n", t.Host, t.Port)
	if n.Host == t.Host {
		fmt.Printf("Elastic config points to %v which is the Front Envoy of MP", n.Host)
		if n.Port == t.Port {
			fmt.Printf("the port %v is also matches (which is only allowed config with Front Envoy for Elastic Search\n", n.Port)
		} else {
			fmt.Printf("*** there seem to be port mismatch (MP - %v and CP - %v while usint the same host\n", t.Port, n.Port)
			fmt.Printf("*** this is invalid config - if MP FrontEnvoy %v is used for Elastic Search - port has to match MP Port %v***\n", t.Host, t.Port)
		}
	} else {
		fmt.Printf("CP plane is configured for direct (not via FrontEnvoy) access to ES via Host - %v and Port - %v \n", n.Host, n.Port)

	}
	fmt.Printf("The protocol used for Elastic Search access is %v\n", n.Protocol)
	if n.SelfSigned {
		fmt.Printf("The ES Endpoint expects CP to trust it via SelfSiged cert from 'es-cert' secret in `istio-system` namespace")
	} else {
		fmt.Printf("In Control Plane config the ES Endpoint marked as one that uses a public CA and can be trusted by any client\n'es-cert' secret in `istio-system` namespace will not be used\n")

	}
}
