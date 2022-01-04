package validator

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

// Codes function takes the HTTP StatusCode and some of tribe knowledge and presents the user
// with the set of suggestions what would the next steps in troubleshooting process
// there are additional checks invokes on as needed basis
// also the attempt to provide short but descriptive output is made here

func Codes(c collector.CPTelemetryStore, oc collector.CPTelemetryStore, r *http.Response, b []byte, mp bool) bool {

	// for Protocol and SelfSigned parameters the message holders are defined
	var pm, sm string

	// output requires some pretty printing and custom function is used here
	p := CustomPrint()

	// the function is provided with Status Code returned by the server to make top-level suggestion

	switch r.StatusCode {

	case 200:

		// while the response from the server is 200, it can be other issues that prevent data from
		// flowing from CP to ES - one of them ES version mismatch that is checked here
		// v is the version returned by the server and m is a bool for the version returned matching the one setup in CP
		m, v := VersionCheck(b, c.Version)

		// version that equals zero means the server response is not what Elastic Search is expect to return
		if v != "0" {
			// comparing oc (original config) with c (current) that produced 200 OK status code
			if oc.Protocol == c.Protocol && oc.SelfSigned == c.SelfSigned && m {
				var cmd [2]string
				cmd[0] = "kubectl -n istio-system scale deployment oap-deployment zipkin --replicas=0"
				cmd[1] = "kubectl -n istio-system scale deployment oap-deployment zipkin --replicas=1"
				fmt.Printf("\n%v\nNo problems detected - your config works as expected (restarting \"oap-deployment\" and \"zipkin\" pods might be required\nThe easiest way to do that is to scale the deployments per:\n%v%v\n%v%v\n", p.Stars, p.Indent, cmd[0], p.Indent, cmd[1])
				fmt.Printf("\nPlease note that because the private key is required to check signature of the tokens, \nif oap is still restarting, then re-issuing the tokens could be the next step\n(to issue new tokens:\n%v- connect to the Kubernetes cluster that runs TSB Management plane\n%v- use tctl command \"tctl install manifest control-plane-secrets --cluster <cluster name>\" to generate the new set of tokens\n%v- save tokens to a file and apply in CP cluster", p.Indent, p.Indent, p.Indent)
			} else {
				// if the oc (original config) has been modified inform user of every change
				fmt.Println(p.Stars)
				if oc.Protocol != c.Protocol {
					fmt.Printf("\nProtocol mismatch found - Current setting - %v Correct setting - %v", oc.Protocol, c.Protocol)
				}

				if oc.SelfSigned != c.SelfSigned {
					fmt.Printf("\n\"SelfSigned\" parameter mismatch found - Current setting - %v Correct setting - %v", oc.SelfSigned, c.SelfSigned)
				}

				// Elastic Search version mismatch is not part of the config and the result of additional check above via VersionCheck function
				if !m {
					fmt.Printf("\n%v\nElastic Search Version mismatch found - Current setting: %v however ES instance returns: %v\n", p.Indent, c.Version, v)
				}

				// providing one liner summary to compare current (c) and original (oc) configs
				fmt.Printf("\n%v\nIn summary:\n", p.Stars)
				fmt.Printf("\nCurrent settings:\n%vHost - %v | Port - %v | Protocol - %v | Selfsigned - %v | Version - %v\n", p.Indent, oc.Host, oc.Port, oc.Protocol, oc.SelfSigned, oc.Version)
				fmt.Printf("\nThe correct settings:\n%vHost - %v | Port - %v | Protocol - %v | Selfsigned - %v | Version - %v\n", p.Indent, c.Host, c.Port, c.Protocol, c.SelfSigned, v)
				// Printing out the properly formatted yaml should also help the operator with correction of CP CRD
				fmt.Printf("\n%v\nAfter corrections the YAML of CP will look like this:", p.Stars)

				// warning that some defaults are not shown in the CRD (doing checks to display only relevant info)
				if c.Protocol == "https" {
					fmt.Printf("\n(Please note that \"protocol: https\" is default settings and it gets removed from the Control Plane CRD)")
					pm = ""
				} else {
					pm = "\n      protocol: " + c.Protocol
				}
				if c.SelfSigned == false {

					fmt.Printf("\n(Please note that \"selfSigned: false\" is default settings and it gets removed from the Control Plane CRD)")
					sm = ""
				} else {
					sm = "\n      selfSigned: " + strconv.FormatBool(c.SelfSigned)
				}

				// proper YAML formatted output
				fmt.Printf("\n  telemetryStore:\n    elastic:\n      host: %v\n      port: %v%v%v\n      version: %v\n", c.Host, c.Port, sm, pm, v)
			}
		} else {
			// if the expected ES response is not received there is no recommendation can be provided, just some additional clues
			fmt.Printf("\n%v\nNot so right...\nWe got the settings that produce correct code but ES is not responding as expected: \n", p.Stars)
			return false
		}

		return true
	case 401:
		// Handling the credentials error response
		fmt.Printf("\nReceived HTTP Code: %v, which means credentials are not correctly specified in \"elastic-credentials\" secret in \"istio-system\"\n", r.StatusCode)
		fmt.Printf("\nPlease fix and rerun if needed\n")
		return true

	case 503:
		// providing additional pointers for 503 Status Code
		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - currently we don't have a solution for you.\n", p.Stars, r.Status, p.Indent)
		if mp {
			fmt.Printf("Below are additional pointers to validated:")
			fmt.Printf("\n%v- check if MP setting working correctly with the Elastic Search", p.Indent)
			fmt.Printf("\n%v- point CP to Elastic search directly\n", p.Indent)
		} else {

		}

	default:
		// additional info is provided when unexpected status code is returned by the server
		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - currently we don't have a solution for you.\n", p.Stars, r.Status, p.Indent)
		if mp {
			fmt.Printf("Pointing CP to Elastic search directly (instead of MP FrontEnvoy) can provide some additional data.\n")
		}

	}

	return false
}
