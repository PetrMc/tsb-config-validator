package main

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/src/collector"
	"github.com/PetrMc/tsb-config-validator/src/validator"
)

func main() {
	cred, conn := collector.CPData()
	validator.Checklist(cred, conn)

	fmt.Printf("Done main module.\n %v\n %v\n", cred, conn)
}
