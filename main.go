package main

import (
	"github.com/PetrMc/tsb-config-validator/src/collector"
	"github.com/PetrMc/tsb-config-validator/src/validator"
)

// main package is currently calling collector package and
// passes the collected data to validator package
func main() {

	// collector.CPData function is called and returned data
	// is stored into the set of variables
	cred, conn, tsb, tokens, err := collector.CPData()

	// validator.Analysis package is called, the collected
	// data from CP is passed to the package to Analyse and
	// output results to the terminal
	if err == nil {
		validator.Analysis(&cred, conn, &tsb, &tokens)
	}
}
