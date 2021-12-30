package validator

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func MPEndPoint(cred *collector.ES, conn collector.CPTelemetryStore, tkn *collector.TSBTokens) {

	p := CustomPrint()

	fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
	r, b := CheckFrontEnvoy(cred, &conn, tkn.Zipkint)

	switch r.StatusCode {

	case 200:

		m, v := VersionCheck(b, conn.Version)
		if m {
			fmt.Printf("\n%v\nResponse status:%v\n%vThe settings seem to be working as it is - no additional checks are done\n", p.Stars, r.Status, p.Indent)
		} else {
			fmt.Printf("\n%v\nResponse status:%v\n", p.Stars, r.Status)
			fmt.Printf("\n%v\nHoever Elastic Search Version mismatch is detected:\nVersion specified in CP is %v\nWhile ES instance returns: %v\n%v", p.Indent, conn.Version, v, p.Stars)
		}
		if conn.SelfSigned {
			fmt.Printf("%v\nFor MP to work correctly please specifiy: --- SelfSignedL: false ---- per below (when \"SelfSigned\" is missing it also means it set to \"false\"", p.Stars)
			fmt.Printf("\n  telemetryStore:\n    elastic:\n      host: %v\n      port: %v\n      selfSigned: false\n      version: %v\n", conn.Host, conn.Port, conn.Version)
			fmt.Println(p.Stars)
		}

		return

	case 401:

		fmt.Printf("\nReceived HTTP Code: %v, which means credentials are not correctly specified in \"elastic-credentials\" secret in \"istio-system\"\n", r.StatusCode)

	default:

		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - the onlys suggestion is to point CP to Elastic search directly\n", p.Stars, r.Status, p.Indent)
		// scope := []string{"https"}
		// BruteForce(cred, conn, scope)

	}
}
