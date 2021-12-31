package collector

import (
	"fmt"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	kubeConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var config *rest.Config

func k8s() *rest.Config {
	var err error
	config, err = kubeConfig.GetConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(123)
	}
	return config
}
