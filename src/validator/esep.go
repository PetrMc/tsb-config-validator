package validator

import (
	"fmt"
	"net/http"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func ESEndpoint(cred *collector.ES, conn collector.CPTelemetryStore) {

	p := CustomPrint()
	fmt.Printf("\nChecking direct connection between CP and MP ")

	r, b := MPCheck(cred, &conn)

	if !Codes(conn, r, b) {
		fmt.Printf("\nResponse status:%v\n%v\nThe settings are not working - the series of tests will try different combination of setting to get to the bottom of the problem\n", r.Status, p.Indent)
		scope := []string{"https", "http"}
		BruteForce(cred, conn, scope)
	} else {
		fmt.Printf("\n%v\nResponse status:%v\n%vThe settings seem to be working as it is - no additional checks are done\n", p.Stars, r.Status, p.Indent)

	}

}

func Codes(c collector.CPTelemetryStore, r *http.Response, b []byte) bool {

	p := CustomPrint()

	switch r.StatusCode {

	case 200:
		m, v := VersionCheck(b, c.Version)

		if v != "0" {
			fmt.Printf("Correct settings:")
			fmt.Printf("\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", c.Host, c.Port, c.Protocol, c.SelfSigned)
			if !m {
				fmt.Printf("Additionally - please make sure ES Version is set to %v\n%v", v, p.Stars)
			}
		} else {
			fmt.Printf("\n%v\nNot so right...\nWe got the settings that produce correct code but ES is not responding as expected: \n", p.Stars)
		}

		return true
	case 401:

		fmt.Printf("\nReceived HTTP Code: %v, which means credentials are not correctly specified in \"elastic-credentials\" secret in \"istio-system\"\n", r.StatusCode)
		fmt.Printf("\nPlease fix and rerun if needed\n")
		return true

	default:
		// fmt.Printf("\nReceived HTTP Code: %v\n", r.StatusCode)
		// fmt.Printf("\nReceived HTTP Code: %v, and have no idea what to do with it. Sorry!\n", r.StatusCode)

	}

	return false
}
