package validator

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func Worker(cred *collector.ES, conn collector.CPTelemetryStore, tkn *collector.TSBTokens, mp bool) {

	// p := CustomPrint()
	oc := conn
	if mp {
		fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
		conn.Protocol = "https"
	} else {
		fmt.Printf("\nChecking direct connection from CP to ElasticSearch\n")
	}

	r, b := ESCheck(cred, &conn, tkn.Zipkint, mp)
	if b == nil {
		conn.Protocol = "https"
		r, b = ESCheck(cred, &conn, tkn.Zipkint, mp)
	}
	if r != nil {
		Codes(conn, oc, r, b, true)
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

	case 200:
		m, v := VersionCheck(b, c.Version)

		if v != "0" {

			if oc.Protocol == c.Protocol && oc.SelfSigned == c.SelfSigned && m {
				fmt.Printf("\n%v\nNo problems detected - your config works as expected", p.Stars)
			} else {
				fmt.Println(p.Stars)
				if oc.Protocol != c.Protocol {
					fmt.Printf("\nProtocol mismatch found - Current setting - %v Correct setting - %v", oc.Protocol, c.Protocol)

				}
				if oc.SelfSigned != c.SelfSigned {
					fmt.Printf("\n\"SelfSigned\" parameter mismatch found - Current setting - %v Correct setting - %v", oc.Protocol, c.Protocol)
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

	default:
		if mp {
			fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - currently we don't have a solution for you.\n", p.Stars, r.Status, p.Indent)
		} else {
			fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - the only suggestion is to point CP to Elastic search directly\n", p.Stars, r.Status, p.Indent)
		}

	}

	return false
}
