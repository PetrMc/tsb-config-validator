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

type TSBTokens struct {
	Oapt, Otelt, Zipkint, Xcpt string
}

func TokensAll(c *kubernetes.Clientset) TSBTokens {
	var tokens = []string{"oap-token", "otel-token", "zipkin-token", "xcp-edge-central-auth-token"}

	tn := TSBTokens{}
	tn.Oapt = Tokens(c, tokens[0])
	tn.Otelt = Tokens(c, tokens[1])
	tn.Zipkint = Tokens(c, tokens[2])
	tn.Xcpt = Tokens(c, tokens[3])

	return tn
}

func Tokens(c *kubernetes.Clientset, t string) string {
	var tkn string
	out, err := c.CoreV1().Secrets("istio-system").Get(context.TODO(), t, metav1.GetOptions{})
	if err == nil {
		tkn = string(out.Data["token"])
		if len(tkn) == 0 {
			tkn = string(out.Data["jwt"])
		}
	}
	return tkn
}
