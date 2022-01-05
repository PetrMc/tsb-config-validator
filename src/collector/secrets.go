package collector

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Secrets function is focused on obtaining credentils for making calls to ElasticSearch
// instance - username, password and CA to validate the ES authenticity
func Secrets(c *kubernetes.Clientset, n string) ES {

	// err to define in hopes will not need to handle those
	var err error
	// s is placeholder for data
	var s ES

	// secret_names has a potential for extension if needed
	var secret_names [2]string
	secret_names[0] = "elastic-credentials"
	secret_names[1] = "es-certs"

	// call is using corev1 library of golang to read the secrets
	es_credentials_secret_mp, err := c.CoreV1().Secrets(n).Get(context.TODO(), secret_names[0], metav1.GetOptions{})
	if err != nil {
		println(err.Error())
	}
	// parsing the secret data
	s.Username = string(es_credentials_secret_mp.Data["username"])
	s.Password = string(es_credentials_secret_mp.Data["password"])

	// quering the es-certs secret
	es_ca_cert_secret_mp, err := c.CoreV1().Secrets(n).Get(context.TODO(), secret_names[1], metav1.GetOptions{})
	if err != nil {
		println(err.Error())
	}

	// assigning certificate data to a variable
	s.Cert = string(es_ca_cert_secret_mp.Data["ca.crt"])
	return s

}

// TokensAll enumerates via all tokens of interests
// calls Tokens (retrival function)
func TokensAll(c *kubernetes.Clientset) TSBTokens {

	// setting the token names to match the canonical names
	var tokens = []string{"oap-token", "otel-token", "zipkin-token", "xcp-edge-central-auth-token"}

	// tn is placeholder for TSBTokens struct
	var tn TSBTokens

	// since enumeration of struct field is complex
	// defining the tokens one by one
	tn.Oapt = TokensGet(c, tokens[0])
	tn.Otelt = TokensGet(c, tokens[1])
	tn.Zipkint = TokensGet(c, tokens[2])
	tn.Xcpt = TokensGet(c, tokens[3])

	return tn
}

// TokensGet function queries k8s cluster
// and returns the token string found
func TokensGet(c *kubernetes.Clientset, t string) string {

	// tkn will store a string of jwt token
	var tkn string

	out, err := c.CoreV1().Secrets("istio-system").Get(context.TODO(), t, metav1.GetOptions{})
	if err == nil {
		// first looking for field "token"
		tkn = string(out.Data["token"])

		// if not found look up "jwt" field (as fields in TSB have either of these two names)
		if len(tkn) == 0 {
			tkn = string(out.Data["jwt"])
		}
	}
	return tkn
}
