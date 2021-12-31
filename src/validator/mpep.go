package validator

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func MPEndPoint(cred *collector.ES, conn collector.CPTelemetryStore, tkn *collector.TSBTokens) {

	// p := CustomPrint()

	fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
	// conn.Protocol for FrontEnvoy can be only http - no point of doing non-https calls
	conn.Protocol = "https"
	r, b := CheckFrontEnvoy(cred, &conn, tkn.Zipkint)

	Codes(conn, r, b, true)

}
