package collector

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type ES struct {
	Username, Password, Cert string
}

func Secrets(c *kubernetes.Clientset, n string) ES {
	s := ES{}
	var secret_names [2]string
	secret_names[0] = "elastic-credentials"
	secret_names[1] = "es-certs"

	es_credentials_secret_mp, err := c.CoreV1().Secrets(n).Get(context.TODO(), secret_names[0], metav1.GetOptions{})
	if err != nil {
		println(err)
	}
	s.Username = string(es_credentials_secret_mp.Data["username"])
	s.Password = string(es_credentials_secret_mp.Data["password"])

	es_ca_cert_secret_mp, err := c.CoreV1().Secrets(n).Get(context.TODO(), secret_names[1], metav1.GetOptions{})
	if err != nil {
		println(err)
	}
	s.Cert = string(es_ca_cert_secret_mp.Data["ca.crt"])
	// fmt.Printf(s.Username, s.Password, s.Cert)
	return s

}

// func GetTokens(c *kubernetes.Clientset, n string) []string {

// 	var es_tokens_cp = []string{"oap-token configured", "otel-token configured", "zipkin-token configured", "xcp-edge-central-auth-token"}
// 	tkn := make([]string, len(es_tokens_cp))

// 	for i, t := range es_tokens_cp {

// 		out, err := c.CoreV1().Secrets(n).Get(context.TODO(), t, metav1.GetOptions{})
// 		if err == nil {
// 			tkn[i] = string(out.Data["jwt"])
// 		}
// 	}

// 	return tkn
// }
