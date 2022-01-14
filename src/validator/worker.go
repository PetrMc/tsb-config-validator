package validator

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

// Worker function is a pivotal part of the validator - it follows the steps to proceed with different checks
// that cover a significant service of the issues that were found in customer and test environments
// the function is being called by Analysis function after doing some basic sanity checks

func Worker(cred *collector.ES, conn collector.CPTelemetryStore, tkn *collector.TSBTokens, mp bool) {

	// The variable p - can also be thought as pretty-print - an attempt to provide more readable output.
	p := CustomPrint()

	// During the checks some of source parametes will be changed - oc (stands for original connection)
	// is used to preserve the original setting and present user with current (aka original) and modified
	// settings (that according to the test performed here should be the correct settings)
	oc := conn

	// Calling MPPrint function for an output that can be viewed as excessive - however it reminds the user
	// what checks exactly will be performed
	MPPrint(mp)

	// First check that is being done is the most important - the setting are used by ESCheck function
	// in attempt to establish the connection between the validator (this code) and Elastic Search instance
	// using exactly the same settings as TSB Control Plane would use. The function returns response (r)
	// and the parsed body (that can have additional details of Elastic Search)
	r, b := ESCheck(cred, &conn, tkn.Zipkint, mp)

	// the useful troubleshooting procedures are also possible with cURL - the below block forms the
	// command line in Linux to allow additional troubleshooting. Currently (can change in future)
	// two headers are used:

	var header [2]string
	// "tsb-route-target" is used to tell FrontEnvoy to act as proxy for Elastic search (technically
	// can be dropped in CP-ES direct connection scenario) - however if present doesn't hurt (left here to simplify the logic)
	header[0] = " -H \"tsb-route-target: elasticsearch\" "
	// x-tetrate-token: is also used by FrontEnvoy to validate the source of the connection
	header[1] = " -H \"x-tetrate-token: " + tkn.Zipkint + "\" "

	// m is a string that contains the whole command output
	m := "curl -u " + cred.Username + ":" + cred.Password + " " + oc.Protocol + "://" + oc.Host + ":" + oc.Port + header[0] + header[1]

	// the Status codes can be analyzed only when respose is received otherwise the only info that can be shared is curl
	// command that has been used (as there is no connectivity)
	if r != nil {
		// Codes function serves as dictionary for the known codes and associated actions to fix those
		Codes(conn, oc, r, b, true)

		// The username and password validity is checked by ESCheck function - those are not valid - Status Code 401 is returned
		// and properly handled by Codes function, additionally - one of the bahavoirs that is addressed here - use of return
		// carriage (\n) in Elastic Search credentail password - if present some calls might succeed when others fail.
		fmt.Printf("\nAnalyzing ES Credentials...\n")
		PasswdCheck(cred)

		// While the signature of tokens can only be validated with Private Key (not available to this code at the moment) - some basic checks
		// can be done such as token expiration date and existence of the tokens
		fmt.Printf("\nChecking tokens presence and expiry date...\n")
		TokenCheck(tkn)
		// currently the code still prints cURL detailed command for testing - if excessive - can be commented or removed
		fmt.Printf("\n%v\nFor debug proposes you can use \"curl\" command per below:\n%v\n", p.Stars, m)
	} else {
		// This is the last resort - when Status Code is not returned - the only step that can be taken is to do basic checks with curl command
		m := "curl -u " + cred.Username + ":" + cred.Password + " " + oc.Protocol + "://" + oc.Host + ":" + oc.Port
		fmt.Printf("\n%v\nNo response received from the server - test with \"curl\" command could provide some networking data:\n%v\n", p.Stars, m)

	}

}
