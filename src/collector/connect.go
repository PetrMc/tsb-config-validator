package collector

import (
	"fmt"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	kubeConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// k8s queries the current terminal context and
// returns config that will be used to fetch CRD data
// and secrets
func k8s() (*rest.Config, error) {

	// err is handling errors
	var err error

	// rest.Config holder
	var config *rest.Config

	config, err = kubeConfig.GetConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	return config, err
}
