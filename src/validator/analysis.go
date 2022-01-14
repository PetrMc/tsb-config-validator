package validator

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

// Analysis is a main function of this package that sorts out the data received from Kubernetes CRD, prints it out and if valid runs the in-depth testing

func Analysis(c *collector.ES, n collector.CPTelemetryStore, t *collector.TSBConf, tkn *collector.TSBTokens) {

	// Custom Print set number of string variables to help with human-readable screen output
	p := CustomPrint()

	// This block displays the info received from Kubernetes CSR as-is
	fmt.Println(p.Stars)
	fmt.Printf("The following data is received for Control plane: --- %v --- \n", t.ClusterName)
	fmt.Printf("%vManagement plane Front Envoy \n%vAddress: %v\n%vPort: %v\n", p.Indent, p.Twoindent, t.Host, p.Twoindent, t.Port)
	fmt.Println(p.Stars)

	// there are two types of Elastic Search communication - (1) via FrontEnvoy (Management Plane component) and (2) directly to the Elastic Search
	// mp variable is used to signify if MP FrontEnvoy if used for Telemetry data - (1) is set to true and (2) is set to false
	var mp bool
	// For (1) - Host value in TSM Management section and Telemetry section should match
	if n.Host == t.Host {
		fmt.Printf("Elastic config points to %v which is the Front Envoy of MP\n", n.Host)
		//if MP and Telemetry host matches the only config that will work if the same port is used - here is check for it
		if n.Port == t.Port {
			fmt.Printf("the port %v is also matches (which is the only allowed config with Front Envoy for Elastic Search)\n", n.Port)
			// Since sanity check succeeds - the variable is set accordingly (FrontEnvoy from Telemetry - mp is true)
			mp = true
		} else {
			// No actions can be taken if the config is invalid from logical point of view - informing operator here and exiting
			fmt.Printf("*** there seem to be port mismatch (MP - %v and CP - %v while using the same host\n", t.Port, n.Port)
			fmt.Printf("*** this is invalid config - if MP FrontEnvoy %v is used for Elastic Search - port has to match MP Port %v***\n", t.Host, t.Port)
			fmt.Printf("Exiting")
			return
		}
	} else {
		// Not too much pre-validation can be done if ElasticSearch is a seaprate instance, the code prints data and proceeds to the detailed checks
		fmt.Printf("CP plane is configured for direct (not via FrontEnvoy) access to ElasticSearch via:\n%vHost: %v\n%vPort %v\n", p.Indent, n.Host, p.Indent, n.Port)
		fmt.Printf("%vProtocol: %v\n", p.Indent, n.Protocol)

		// While all other parameters are selfexplanable "selfSigned" might require a bit of additional human readable output
		// SSPrint function is used to do that and it only displayed for https connection as http will not use the certs
		if n.Protocol != "http" {
			SSPrint(n.SelfSigned)
		}

		// The variable is set accordingly (No FrontEnvoy is used for ElasticSearch - mp is false)
		mp = false
	}

	// if the source configuration succeed we proceed with the detailed checks by passing Credentials, Telemetry, Tokens and boolean on FrontEnvoy use 
	// Worker function that performs deeper analysis
	Worker(c, n, tkn, mp)

	fmt.Println(p.Stars)

	return
}


