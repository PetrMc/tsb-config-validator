package validator

import (
	"fmt"
	"strings"

	tsize "github.com/kopoli/go-terminal-size"
)

// PrintFields struct is used to define the parameters that are used for human-friendly output
type PrintFields struct {
	Width                    int
	Stars, Indent, Twoindent string
}

// CustomPrint function creates set of variables that are handy to use when formatting the output
func CustomPrint() PrintFields {
	s, _ := tsize.GetSize()
	p := PrintFields{
		// to print the separater as full width terminal length - require to know the terminal length
		Width: s.Width,
		// currently line of stars is used to separate the portions of the screen output
		Stars: "\n" + strings.Repeat("*", s.Width) + "\n",
		// Indent are used to properly format yaml-like outputs
		Indent:    strings.Repeat(" ", 2),
		Twoindent: strings.Repeat(" ", 4),
	}

	return p
}

// SSPrint function is a simple print out of the explanatory details of "selfSinged" setting in CP CRD
func SSPrint(ss bool) {
	// utilizing the default screenformatting variables
	p := CustomPrint()

	if ss {
		fmt.Printf("%vThe ES Endpoint expects CP to trust it via SelfSigned cert from 'es-cert' secret in `istio-system` namespace\n", p.Indent)
	} else {
		fmt.Printf("%vIn Control Plane config the ES Endpoint marked as one that uses a public CA and can be trusted by any client\n%v'es-cert' secret in `istio-system` namespace will not be used\n", p.Indent, p.Indent)
	}
}

// MP Print function reminds the user of the config that is being tested

func MPPrint(mp bool) {

	if mp {
		fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
	} else {
		fmt.Printf("\nChecking direct connection from CP to ElasticSearch\n")
	}
}
