package collector

import (
	"fmt"

	"github.com/PetrMc/tsb-config-validator/api/v1alpha1/controlplane"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CPData is the main function that is build to run series of
// requests and return all the data that is found in K8s cluster
// that hosts TSB ControlPlane
func CPData() (ES, CPTelemetryStore, TSBConf, TSBTokens, error) {

	// define the standard golang error
	var err error

	// clientset is required to collect secrets from k8s cluster
	var clientset *kubernetes.Clientset

	// creds holds username and password for ElasticSearch
	var creds ES

	// tokens stores TSB related tokens (currently four of those)
	var tokens TSBTokens

	// conn are for values from TelemetryStore portion of
	// TSB ControlPlane CRD
	var conn CPTelemetryStore

	// tsb stores managementPlane portion of TSB
	// Controlplane CRD
	var tsb TSBConf

	// cp is to store TSB Controlplane CRD
	var cp *controlplane.ControlPlane

	// k8s is used to store the populated config
	var k8scfg *K8sClient

	// config is initialized and will be used to get the rest config
	var config *rest.Config
	config, err = k8s()
	if err != nil {
		fmt.Println(err.Error())
		// no reason to proceed if k8s config causes error
		goto end
	}
	

	k8scfg, err = NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
		// no reason to proceed if k8s config causes error
		goto end
	}

	// cp gets TSB CP CRD variables stored
	cp, err = k8scfg.CP("istio-system").Get("controlplane")
	if err != nil {
		fmt.Println(err.Error())
		// no reason to proceed if k8s config causes error
		goto end
	}

	// assigning the required values to variables using
	// parameters function
	conn, tsb = parameters(cp)

	// there is slightly different config that is required to collect secrets
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
		// no reason to proceed if k8s config causes error
		goto end
	}

	// if all previous successful - then querying the specific
	// K8s elements secrets in this case
	creds = Secrets(clientset, "istio-system")
	tokens = TokensAll(clientset)

	// end label is used when error occurs to jump to the end
end:
	return creds, conn, tsb, tokens, err

}

// parameters is used to receive CP CRD and parse it to specific variables
// by calling additional functions
func parameters(c *controlplane.ControlPlane) (CPTelemetryStore, TSBConf) {

	// telemetryStore values are assigned using telemetry function
	var j CPTelemetryStore = telemetry(c)

	// managementPlane values are assigned by calling tsb function
	var k TSBConf = tsb(c)

	return j, k
}

// telemetry function assigns values received from TSB CRD
// and parses them to CPTelemetryStore struct
func telemetry(c *controlplane.ControlPlane) CPTelemetryStore {

	// b is a holder that is being initialized here
	var b CPTelemetryStore

	// Host and Port are self-explanatory
	b.Host = c.Spec.TM.Elastic.Host
	// Port in the struct here requires string type and therefor
	// being converted
	b.Port = fmt.Sprint(c.Spec.TM.Elastic.Port)

	// https is the default setting and it's being dropped (if set to default)
	// for the validation we need to have the value assigned
	if c.Spec.TM.Elastic.Protocol == "" {
		b.Protocol = "https"
	} else {
		b.Protocol = c.Spec.TM.Elastic.Protocol
	}

	// SelfSigned is a boolean value (fortunately if not specified
	// - means default value is set - which is false - so no conversion
	// is required for either case)

	b.SelfSigned = c.Spec.TM.Elastic.SelfSigned

	// Version requires conversion to string type
	b.Version = fmt.Sprint(c.Spec.TM.Elastic.Version)

	return b
}

// tsb function is to assign managemenPlane values to
// TSBConf struct
func tsb(c *controlplane.ControlPlane) TSBConf {

	// TSBConf requires variable
	var b TSBConf

	b.Host = c.Spec.MP.Host

	// Port is converted to string
	b.Port = fmt.Sprint(c.Spec.MP.Port)
	b.ClusterName = c.Spec.MP.ClusterName

	return b
}
