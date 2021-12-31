package validator

import (
	"fmt"
	"net/http"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func Worker(cred *collector.ES, conn collector.CPTelemetryStore) {

	p := CustomPrint()
	fmt.Printf("\nChecking direct connection from CP to MP\n")

	r, b := MPCheck(cred, &conn)

	if !Codes(conn, r, b, false) {
		fmt.Printf("\nResponse status:%v\n%v\nThe settings are not working - trying to identify the source of the issue\n", r.Status, p.Indent)
		Checks(cred, conn)

	} else {
		fmt.Printf("\n%v\nResponse status:%v\n%vThe settings seem to be working as it is - no additional checks are done\n", p.Stars, r.Status, p.Indent)

	}

}

func Checks(c *collector.ES, n collector.CPTelemetryStore) {
	if n.Protocol == "https" {
		if n.SelfSigned {
			if len(c.Cert) == 0 {
				fmt.Printf("\nThe self-signed is set to \"%v\". However \"es-cert\" is not received from \"istio-system\" namespace\nPlease create the secret...\n", n.SelfSigned)
				CertCheck(n.Host, n.Port, n.SelfSigned, c.Cert, false)
				return
			} else {
				CertCheck(n.Host, n.Port, n.SelfSigned, c.Cert, true)
			}
		}
	}

}

func Codes(c collector.CPTelemetryStore, r *http.Response, b []byte, mp bool) bool {

	p := CustomPrint()

	switch r.StatusCode {

	case 200:
		m, v := VersionCheck(b, c.Version)

		if v != "0" {

			fmt.Printf("Correct settings:")
			fmt.Printf("\n  telemetryStore:\n    elastic:\n      host: %v\n      port: %v\n      selfSigned: %v\n      protocol: %v\n      version: %v\n", c.Host, c.Port, c.SelfSigned, c.Protocol, c.Version)
			// fmt.Printf("\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", c.Host, c.Port, c.Protocol, c.SelfSigned)
			if !m {
				// fmt.Printf("Additionally - please make sure ES Version is set to %v (currently shows %v)\n", v, c.Version)
				fmt.Printf("\n%v\nHoever Elastic Search Version mismatch is detected:\nVersion specified in CP is %v\nWhile ES instance returns: %v\n%v", p.Indent, c.Version, v, p.Stars)

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
