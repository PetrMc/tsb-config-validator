package collector

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1"
	"k8s.io/client-go/kubernetes"
)

// config variable stores current k8s context of the user environment

func CPData() (ES, CPTelemetryStore) {
	config = k8s()
	cpcs, err := NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(123)
	}

	cp, err := cpcs.CP("istio-system").Get("controlplane")
	if err != nil {

		fmt.Println(err)
		os.Exit(123)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(123)
	}

	creds := Secrets(clientset, "istio-system")

	// var conn CPTelemetryStore
	conn := parameters(cp)

	// es_username, es_password, es_ca_cert := Secrets(clientSet, "istio-system")
	// fmt.Println(es_username, es_password, es_ca_cert)

	return creds, conn

}

func parameters(c *v1alpha1.ControlPlane) CPTelemetryStore {
	b := CPTelemetryStore{}
	ts := make(map[string]string)

	st := fmt.Sprintf("%v", c)
	re := regexp.MustCompile(`"telemetryStore":{"elastic":\{(.*?)\}`)
	ar := strings.Split(re.FindStringSubmatch(st)[1], ",")

	for i := range ar {
		kv := strings.Split(ar[i], ":")
		ts[strings.Trim(kv[0], "\"")] = strings.Trim(kv[1], "\"")

	}

	b.Host = ts["host"]
	b.Port = ts["port"]
	b.Version = ts["version"]
	b.SSL, _ = strconv.ParseBool(ts["selfSigned"])

	return b
}
