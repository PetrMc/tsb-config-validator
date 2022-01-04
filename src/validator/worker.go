package validator

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func Worker(cred *collector.ES, conn collector.CPTelemetryStore, tkn *collector.TSBTokens, mp bool) {

	p := CustomPrint()
	oc := conn
	if mp {
		// if conn.Protocol == "http" {
		// 	conn.Protocol = "https"
		// 	fmt.Printf("\n\"protocol: http\" will not work with FrontEnvoy - testing with \"https\" instead\n")
		// }

		fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
	} else {
		fmt.Printf("\nChecking direct connection from CP to ElasticSearch\n")
	}

	r, b := ESCheck(cred, &conn, tkn.Zipkint, mp)
	// if b == nil {
	// 	conn.Protocol = "https"
	// 	r, b = ESCheck(cred, &conn, tkn.Zipkint, mp)
	// }
	// fmt.Printf(r.Status)
	var header [2]string
	header[0] = " -H \"tsb-route-target: elasticsearch\" "
	header[1] = " -H \"x-tetrate-token: " + tkn.Zipkint + "\" "

	m := "curl -u " + cred.Username + ":" + cred.Password + " " + oc.Protocol + "://" + oc.Host + ":" + oc.Port + header[0] + header[1]

	if r != nil {
		Codes(conn, oc, r, b, true)
		fmt.Printf("\n%v\nFor debug proposes you can use \"curl\" command per below:\n%v\n", p.Stars, m)
	} else {
		m := "curl -u " + cred.Username + ":" + cred.Password + " " + oc.Protocol + "://" + oc.Host + ":" + oc.Port
		fmt.Printf("\n%v\nNo reponse received from the server - test with \"curl\" command could provide some networking data:\n%v\n", p.Stars, m)

	}

}

// func Checks(c *collector.ES, n collector.CPTelemetryStore) {
// 	if n.Protocol == "https" {
// 		if n.SelfSigned {
// 			if len(c.Cert) == 0 {
// 				fmt.Printf("\nThe self-signed is set to \"%v\". However \"es-cert\" is not received from \"istio-system\" namespace\nPlease create the secret...\n", n.SelfSigned)
// 				CertCheck(n.Host, n.Port, n.SelfSigned, c.Cert, false)
// 				return
// 			} else {
// 				CertCheck(n.Host, n.Port, n.SelfSigned, c.Cert, true)
// 			}
// 		}
// 	}

// }

func Codes(c collector.CPTelemetryStore, oc collector.CPTelemetryStore, r *http.Response, b []byte, mp bool) bool {
	var pm, sm string
	p := CustomPrint()
	// fmt.Println(r.StatusCode)
	switch r.StatusCode {

	// case 0:

	// 	fmt.Printf("\n%v\nNo response from the server. Please review host/port number and firewall settings before trying again\n", p.Stars)
	case 200:
		m, v := VersionCheck(b, c.Version)

		if v != "0" {

			if oc.Protocol == c.Protocol && oc.SelfSigned == c.SelfSigned && m {
				fmt.Printf("\n%v\nNo problems detected - your config works as expected (restarting \"oap-deployment\" and \"zipkin\" pods might be required", p.Stars)
				fmt.Printf("\nPlease note that because the private key is required to check signature of the tokens, \nif oap is still restarting, then re-issuing the tokens could be the next step\n(to issue new tokens:\n%v- connect to the kubernes cluster that runs TSB Management plane\n%v- use tctl command \"tctl install manifest control-plane-secrets --cluster <cluster name>\" to generate the new set of tokens\n%v- save tokens to a file and apply in CP cluster", p.Indent, p.Indent, p.Indent)
			} else {
				fmt.Println(p.Stars)
				if oc.Protocol != c.Protocol {
					fmt.Printf("\nProtocol mismatch found - Current setting - %v Correct setting - %v", oc.Protocol, c.Protocol)

				}
				if oc.SelfSigned != c.SelfSigned {
					fmt.Printf("\n\"SelfSigned\" parameter mismatch found - Current setting - %v Correct setting - %v", oc.SelfSigned, c.SelfSigned)
				}
				if !m {
					fmt.Printf("\n%v\nElastic Search Version mismatch found - Current setting: %v however ES instance returns: %v\n", p.Indent, c.Version, v)
				}
				fmt.Printf("\n%v\nIn summary:\n", p.Stars)
				fmt.Printf("\nCurrent settings:\n%vHost - %v | Port - %v | Protocol - %v | Selfsigned - %v | Version - %v\n", p.Indent, oc.Host, oc.Port, oc.Protocol, oc.SelfSigned, oc.Version)
				fmt.Printf("\nThe correct settings:\n%vHost - %v | Port - %v | Protocol - %v | Selfsigned - %v | Version - %v\n", p.Indent, c.Host, c.Port, c.Protocol, c.SelfSigned, v)

				fmt.Printf("\n%v\nAfter corrections the YAML of CP will look like this:", p.Stars)

				if c.Protocol == "https" {
					fmt.Printf("\n(Please note that \"protocol: https\" is default settings and it gets removed from the Control Plane CRD)")
					pm = ""
				} else {
					pm = "\n      protocol: " + c.Protocol
				}
				if c.SelfSigned == false {

					fmt.Printf("\n(Please note that \"selfSined: false\" is default settings and it gets removed from the Control Plane CRD)")
					sm = ""
				} else {
					sm = "\n      selfSigned: " + strconv.FormatBool(c.SelfSigned)
				}

				// fmt.Printf("\n  telemetryStore:\n    elastic:\n      host: %v\n      port: %v\n      selfSigned: %v\n      protocol: %v\n      version: %v\n", c.Host, c.Port, ps, c.Protocol, c.Version)
				fmt.Printf("\n  telemetryStore:\n    elastic:\n      host: %v\n      port: %v%v%v\n      version: %v\n", c.Host, c.Port, sm, pm, v)
			}
		} else {
			fmt.Printf("\n%v\nNot so right...\nWe got the settings that produce correct code but ES is not responding as expected: \n", p.Stars)
			return false
		}

		return true
	case 401:

		fmt.Printf("\nReceived HTTP Code: %v, which means credentials are not correctly specified in \"elastic-credentials\" secret in \"istio-system\"\n", r.StatusCode)
		fmt.Printf("\nPlease fix and rerun if needed\n")
		return true

	case 503:
		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - currently we don't have a solution for you.\n", p.Stars, r.Status, p.Indent)
		if mp {
			fmt.Printf("Below are additional pointers to validated:")
			fmt.Printf("\n%v- check if MP setting working correctly with the Elastic Search", p.Indent)
			fmt.Printf("\n%v- point CP to Elastic search directly\n", p.Indent)
		} else {

		}

	default:

		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - currently we don't have a solution for you.\n", p.Stars, r.Status, p.Indent)
		if mp {
			fmt.Printf("Pointing CP to Elastic search directly (instead of MP FrontEnvoy) can provide some additional data.\n")
		}

	}

	return false
}
