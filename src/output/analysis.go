package output

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func PrintCP(c *collector.ES, n *collector.CPTelemetryStore, t *collector.TSBConf) {

	p := CustomPrint()

	fmt.Println(p.Stars)
	fmt.Printf("The following data is received for Control plane: --- %v --- \n", t.ClusterName)
	fmt.Printf("%vManagement plane Front Envoy \n%vAddress: %v\n%vPort: %v\n", p.Indent, p.Twoindent, t.Host, p.Twoindent, t.Port)
	fmt.Println(p.Stars)
	if n.Host == t.Host {
		fmt.Printf("Elastic config points to %v which is the Front Envoy of MP", n.Host)
		if n.Port == t.Port {
			fmt.Printf("the port %v is also matches (which is only allowed config with Front Envoy for Elastic Search\n", n.Port)
		} else {
			fmt.Printf("*** there seem to be port mismatch (MP - %v and CP - %v while using the same host\n", t.Port, n.Port)
			fmt.Printf("*** this is invalid config - if MP FrontEnvoy %v is used for Elastic Search - port has to match MP Port %v***\n", t.Host, t.Port)
		}
	} else {
		fmt.Printf("CP plane is configured for direct (not via FrontEnvoy) access to ElasticSearch via:\n%vHost: %v\n%vPort %v\n", p.Indent, n.Host, p.Indent, n.Port)

	}
	fmt.Printf("%vProtocol: %v\n", p.Indent, n.Protocol)
	if n.SelfSigned {
		fmt.Printf("%vThe ES Endpoint expects CP to trust it via SelfSiged cert from 'es-cert' secret in `istio-system` namespace\n", p.Indent)
	} else {
		fmt.Printf("%vIn Control Plane config the ES Endpoint marked as one that uses a public CA and can be trusted by any client\n%v'es-cert' secret in `istio-system` namespace will not be used\n", p.Indent, p.Indent)

	}
	fmt.Println(p.Stars)
}
