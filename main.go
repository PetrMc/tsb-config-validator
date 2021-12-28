package main

import (
	"github.com/PetrMc/tsb-config-validator/src/collector"
	"github.com/PetrMc/tsb-config-validator/src/output"
	"github.com/PetrMc/tsb-config-validator/src/validator"
)

func main() {
	cred, conn, tsb, tokens := collector.CPData()

	fe := output.PrintCP(&cred, &conn, &tsb)

	validator.Checklist(&cred, conn, tokens, fe)

	// fmt.Printf("Done main module.\n tsb port: %v\ncred username: %v pass: %v\n conn: %v\n", tsb, cred.Username, cred.Password, conn)

}
