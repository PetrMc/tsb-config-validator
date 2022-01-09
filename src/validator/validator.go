package validator

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

// ESCheck function analyzes the provided settings and tries to validate those that believed to be correct
// by initiating http(s) call to the MP

func ESCheck(cr *collector.ES, c *collector.CPTelemetryStore, t string, mp bool) (*http.Response, []byte) {

	// resp, tc and pool are initialized here to accommodate a proper http call
	var resp *http.Response
	var tc *tls.Config
	pool := x509.NewCertPool()

	// b is used to store body response
	var b []byte

	// The function interacts with the operator and requires readable output - will use the function
	// to help with it
	p := CustomPrint()

	fmt.Printf("\nEstablishing connection... ")

	// tr is initialized with default parameters and
	// additional setting added as evaluation of CP Parameters goes
	tr := http.DefaultTransport.(*http.Transport).Clone()

	// FrontEnvoy always runs on Secure port, testing HTTP connection is dangerous
	// as Redirect 301 might happen for unsecure connection which can be ok with
	// UI but not properly handled for ElasticSearch - in short only HTTPS is used for
	// ElasticSearch communication via FrontEnvoy

	if !mp {
		fmt.Printf("Trying PLAIN-TEXT connection to ES...\n")
		// to perform http test the c.Protocol needs to be set to http
		// if not successful it will be changed to https later in this function
		c.Protocol = "http"
		resp, b = ESDial(cr, c, t, mp, tr)

		if b == nil {
			fmt.Printf("Can't establish plain HTTP connection, will try HTTPS")
		} else {
			// if successful no further checks is performed
			return resp, b
		}
	} else {
		fmt.Printf("MP will only work with HTTPS skipping HTTP\n")
	}

	// the rest of the checks is done using HTTPS that is set here explicitly
	c.Protocol = "https"

	fmt.Printf("\nTrying ENCRYPTED connection to ES\n")
	// Before making HTTPs call - the server end cert needs to be queried
	// most of the decisions below based on the cert presented
	// SRVCert function makes TLS TCP call, not http one.
	srv, srvcert := SRVCert(c.Host, c.Port)

	// if there is no server side cert presented - there is no reason to continue
	// this function will return nil, nil and the human operator needs to make the decisions
	// based on high level config (i.e. if ElasticSearch reachable at all)
	if srv {
		// if Publicly verifiable server cert presented - no need for the cert stored locally in CP Kubernetes cluster (es-certs in istio-system)
		if IsPublic(srvcert) {
			if c.SelfSigned {
				fmt.Printf("\nThe server presents the publicly signed certificate - the current CP settings states that \"selfSigned: true\" while should be set to \"false\" (testing with \"false\" setting)\n")
				c.SelfSigned = false
			}
		} else {
			if !c.SelfSigned {
				fmt.Printf("\nThe server presents the self-signed certificate - the current CP settings states that \"selfSigned: false\" while should be set to \"true\" (testing with \"true\" setting)\n")
				c.SelfSigned = true
			}
			// for Self-signed certificate - it needs to be analyzed
			// first question - is there a cert in es-certs / istio-system namespace?
			if len(cr.Cert) == 0 {
				// is the server cert has CA in its chain?
				if IsCA(srvcert) {
					// since there is no certificate in Kubernetes secret
					// saving CA to the file - the operator can create the secret based on it.

					fn := "/tmp/ca.crt"
					CASave(srvcert, fn)

					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=" + fn + "\n"
					fmt.Printf("\nThere in no certificate stored in \"es-certs\" in \"istio-system\" however the server doesn't has CA cert in its chain\nPlease create it in you CP cluster per:\n%v", cmd)
				} else {

					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=<file that contains CA cert>\n"
					fmt.Printf("\nThere in no certificate stored in \"es-certs\" in \"istio-system\" and the server doesn't have CA cert in its chain\nYou have to obtain the CA cert *manually* and add it to the secret\nper:\n%v", cmd)
				}
				// if can't obtain the cert for self-signed server cert - there is no point
				// of continue testing until the secret is createde and contains the correct info.
				return nil, nil
			} else {
				// Handling the cases when the CA cert is presented by the server
				if IsCA(srvcert) {
					fmt.Printf("\nThe server has CA cert in its chain")
					// First compare with kubernetes secret
					if IsMatch(srvcert, cr.Cert) {
						fmt.Printf(" and there is a matching certificate stored in \"es-certs\" in \"istio-system\" \nWill try to call the MP using the settings")
					} else {
						// if not save to the file and ask operator to apply.
						fn := "/tmp/ca.crt"
						cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=" + fn + "\n"
						fmt.Printf(" however the certificate stored in \"es-certs\" in \"istio-system\" doesn't match that chain\nPlease create it in you CP cluster per:\n%v", cmd)
						CASave(srvcert, fn)
						// no reason to test until the secret is created
						return nil, nil
					}
				} else {
					// Test if the server chain doesn't have CA cert but there is a cert in Kubernetes secret
					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=<file that contains CA cert>\n"
					fmt.Printf("\nValidating connection using the certificate stored in \"es-certs\" in \"istio-system\" - if the settings don't work - you will have to obtain the CA cert *manually* and add it to the secret\nper:\n%v", cmd)
				}
			}
		}
	} else {
		// if the server doesn't return the certificate - returning as there is no more tests possible
		// network connectivity needs to be validated first
		return nil, nil
	}

	// in SelfSigned scenario the stored CA is being added to the request.
	if c.SelfSigned {
		tc = &tls.Config{RootCAs: pool}
		if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
			fmt.Println("Failed to append cert")
			tc = &tls.Config{RootCAs: pool}
		}
	}

	// tr (transport is finalized here)
	tr = &http.Transport{TLSClientConfig: tc}
	// the function will return data from the test http call
	// run by data generated here and executed in ESDial function
	return ESDial(cr, c, t, mp, tr)
}

// ESDial uses net/http client to make the call to Elastic search
// it returns the response collected and the output parsed body
// without any futher validation

func ESDial(cr *collector.ES, c *collector.CPTelemetryStore, t string, mp bool, tr *http.Transport) (*http.Response, []byte) {

	// req, resp are are initialized here to accommodate a proper http(c) call

	var req *http.Request
	var resp *http.Response

	//path and err are self-explanatory
	var path string
	var err error
	// b is used to store body response

	var b []byte

	client := &http.Client{Transport: tr}

	path = c.Protocol + "://" + c.Host + ":" + c.Port

	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	// to allow ElasticSearch respond to the function proper credentials are required
	// the credentials are stored in elastic-credentials secret in istio-system
	// here this values are added to the http request
	req.SetBasicAuth(cr.Username, cr.Password)

	// Headers can be added to the request in the following format:
	// // req.Header.Set("name", "value")

	// "tsb-route-target" is used to tell FrontEnvoy to act as proxy for Elastic search (technically
	// can be dropped in CP-ES direct connection scenario) - however if present doesn't hurt (left here to simplify the logic)
	req.Header.Set("tsb-route-target", "elasticsearch")

	// x-tetrate-token: is also used by FrontEnvoy to validate the source of the connection	//
	req.Header.Set("x-tetrate-token", t)

	// executing the call
	resp, err = client.Do(req)

	// handling error - returning as Response Status and supplying empty body of the response
	if err != nil {
		resp = new(http.Response)
		resp.Status = err.Error()
		b = nil

	} else {
		// also function processes and saves the body of the request
		// that can be used later to gather additional details around
		// ElasticSearch instance
		b, err = io.ReadAll(resp.Body)

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return resp, b
}

func VersionCheck(b []byte, v string) (bool, string) {

	// ElasticSearch responds with JSON containing
	// the information about the instance
	// currently - the only information of interest is
	// version - the corresponding struct is defined

	data := ESResponse{}

	// err will mean there is an unexpected response from ES Server
	// which will also signal the response from the server is unexpected
	// no others service would be successful connecting using this parameters
	err := json.Unmarshal([]byte(b), &data)
	if err != nil {
		fmt.Println(err.Error())
		return false, "0"
	}

	// fetching Elastic Search version number and
	// comparing with CP CRD settings

	if data.Version.Number[0:1] == v {
		return true, data.Version.Number[0:1]
	} else {
		return false, data.Version.Number[0:1]
	}

}

// ESResponse is a struct for JSON returned by ElasticSearch
type ESResponse struct {

	// Version ESVersion
	Version struct {
		Number string
	}
}

// PasswdCheck function makes sure there is no return carriage "\n"
// characters in Base64 coded credentials (elastic-credentials secret
// in istio-system namespace)
func PasswdCheck(cr *collector.ES) {

	// Using internal function to accommodate readable output
	p := CustomPrint()

	// checking if credentials have values

	if len(cr.Password) == 0 || len(cr.Username) == 0 {
		fmt.Printf("\n%v\nNot able to retieve username or password from \"elastic-search secret\" in \"istio-system\" namaspace. Please check if the secret exists\n", p.Stars)
		return
	}

	// comparing the username received (origu) with one that has "\n" removed (modu)
	origu := base64.StdEncoding.EncodeToString([]byte(cr.Username))
	modu := base64.StdEncoding.EncodeToString([]byte(strings.Replace(cr.Username, "\n", "", -1)))

	if origu != modu {
		fmt.Printf("\n%v\nUsername seems to have return carriage \"\\n\" in it. \nPlease update \"elastic-search secret\" in \"istio-system\" namaspace.\nUsername should be %v (currently it returns %v)\n", p.Stars, modu, origu)
	}

	// comparing the password received (origp) with one that has "\n" removed (modp)
	origp := base64.StdEncoding.EncodeToString([]byte(cr.Password))
	modp := base64.StdEncoding.EncodeToString([]byte(strings.Replace(cr.Password, "\n", "", -1)))

	if origp != modp {
		fmt.Printf("\n%v\nPassword seems to have return carriage \"\\n\" in it. \nPlease update \"elasitc-search secret\" in \"istio-system\" namaspace.\nPassword should be %v (currently it returns %v)\n", p.Stars, modp, origp)
	}

}
