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

// func BruteForce(cr *collector.ES, c collector.CPTelemetryStore, prt []string) {

// 	var r *http.Response
// 	var b []byte
// 	tconn := c

// 	p := CustomPrint()

// 	// prt := []string{"http", "https"}
// 	crt := []bool{true, false}

// 	for _, i := range prt {
// 		for _, ss := range crt {

// 			if i != c.Protocol || ss != c.SelfSigned {

// 				tconn.Protocol = i
// 				tconn.SelfSigned = ss
// 				fmt.Printf("\n%v\nTrying the following combination:\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", p.Stars, tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
// 				r, b = MPCheck(cr, &tconn)
// 				fmt.Printf("Response status:%v\n", r.Status)
// 			}
// 			if Codes(tconn, r, b, false) {
// 				return
// 			}
// 			tconn = c

// 		}
// 	}
// }

func PasswdCheck(cr *collector.ES) {

	p := CustomPrint()

	origu := base64.StdEncoding.EncodeToString([]byte(cr.Username))

	if len(cr.Password) == 0 || len(cr.Username) == 0 {
		fmt.Printf("\n%v\nNot able to retieve username or password from \"elasitc-search secret\" in \"istio-system\" namaspace. Please check if the secret exists\n", p.Stars)
		return
	}

	modu := base64.StdEncoding.EncodeToString([]byte(strings.Replace(cr.Username, "\n", "", -1)))
	if origu != modu {

		fmt.Printf("\n%v\nUsername seems to have return cariage \"\\n\" in it. \nPlease update \"elasitc-search secret\" in \"istio-system\" namaspace.\nUsername should be %v (currently it returns %v)\n", p.Stars, modu, origu)

	}

	origp := base64.StdEncoding.EncodeToString([]byte(cr.Password))

	modp := base64.StdEncoding.EncodeToString([]byte(strings.Replace(cr.Password, "\n", "", -1)))
	if origp != modp {

		fmt.Printf("\n%v\nPassword seems to have return cariage \"\\n\" in it. \nPlease update \"elasitc-search secret\" in \"istio-system\" namaspace.\nPassword should be %v (currently it returns %v)\n", p.Stars, modp, origp)

	}

}

func ESCheck(cr *collector.ES, c *collector.CPTelemetryStore, t string, mp bool) (*http.Response, []byte) {
	var resp *http.Response
	var tc *tls.Config
	var b []byte

	pool := x509.NewCertPool()
	p := CustomPrint()

	fmt.Printf("\nEstablishing connection... ")

	tr := http.DefaultTransport.(*http.Transport).Clone()

	if !mp {
		fmt.Printf("Trying PLAIN-TEXT connection to ES...\n")
		c.Protocol = "http"
		resp, b = ESDial(cr, c, t, mp, tr)
		fmt.Println(resp.Status)
		if b == nil {
			fmt.Printf("Can't establish plain HTTP connection, will try HTTPS")
			c.Protocol = "https"
		} else {
			return resp, b
		}
	} else {
		fmt.Printf("MP will only work with HTTPS skipping HTTP\n")
		c.Protocol = "https"

	}

	fmt.Printf("\nTrying ENCRYPTED connection to ES\n")
	srv, srvcert := SRVCert(c.Host, c.Port)
	if srv {
		if IsPulic(srvcert) {
			// is es-cert needed?
			if c.SelfSigned {
				fmt.Printf("\nThe server presents the publicly signed certificate - the current CP settings states that \"selfSigned: true\" while should be set to \"false\" (testing with \"false\" setting)\n")
				c.SelfSigned = false
			}
			// tc = &tls.Config{InsecureSkipVerify: true}
			// tr := http.DefaultTransport.(*http.Transport).Clone()
			// resp, b = ESDial(cr, c, t, mp, tr)

			// return resp, b

		} else {

			if !c.SelfSigned {
				fmt.Printf("\nThe server presents the self-signed certificate - the current CP settings states that \"selfSigned: false\" while should be set to \"true\" (testing with \"true\" setting)\n")
				c.SelfSigned = true
			}
			if len(cr.Cert) == 0 {
				if IsCA(srvcert) {
					fn := "/tmp/ca.crt"
					CASave(srvcert, fn)

					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=" + fn + "\n"
					fmt.Printf("\nThere in no certificate stored in \"es-certs\" in \"istio-system\" however the server doesn't has CA cert in its chain\nPlease create it in you CP cluster per:\n%v", cmd)
				} else {
					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=<file that contains CA cert>\n"
					fmt.Printf("\nThere in no certificate stored in \"es-certs\" in \"istio-system\" and the server doesn't have CA cert in its chain\nYou have to obtain the CA cert *manually* and add it to the secret\nper:\n%v", cmd)
				}
				return nil, nil
			} else {
				if IsCA(srvcert) {
					fmt.Printf("\nThe server has CA cert in its chain")

					if IsMatch(srvcert, cr.Cert) {
						fmt.Printf(" and there in a matching certificate stored in \"es-certs\" in \"istio-system\" \nWill try to call the MP using the settings")
					} else {
						fn := "/tmp/ca.crt"
						cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=" + fn + "\n"
						fmt.Printf(" however the certificate stored in \"es-certs\" in \"istio-system\" doesn't match that chain\nPlease create it in you CP cluster per:\n%v", cmd)
						CASave(srvcert, fn)
						return nil, nil
					}
				} else {
					cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=<file that contains CA cert>\n"
					fmt.Printf("\nValidating connection using the certificate stored in \"es-certs\" in \"istio-system\" - if the settings don't work - you will have to obtain the CA cert *manually* and add it to the secret\nper:\n%v", cmd)
				}
			}
		}
	} else {
		return nil, nil
	}
	tc = &tls.Config{InsecureSkipVerify: true}

	if c.SelfSigned {
		tc = &tls.Config{RootCAs: pool}
		if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
			fmt.Println("Failed to append cert")
			tc = &tls.Config{RootCAs: pool}
		}
	}

	tr = &http.Transport{TLSClientConfig: tc}
	return ESDial(cr, c, t, mp, tr)
}

// func ESCheckOld(cr *collector.ES, c *collector.CPTelemetryStore, t string, mp bool) (*http.Response, []byte) {
// 	var resp *http.Response
// 	var tc *tls.Config
// 	var b []byte

// 	pool := x509.NewCertPool()
// 	p := CustomPrint()

// 	fmt.Printf("\nEstablishing connection... ")
// 	// if len(cr.Cert) != 0 {
// 	// 	if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
// 	// 		fmt.Println("Failed to append cert")
// 	// 		tc = &tls.Config{RootCAs: pool}
// 	// 	}
// 	// } else {
// 	// 	fmt.Printf("\"es-certs\" doesn't have the expected certificate (or the secret doesn't exist at all)... Trying connectivity using Public PKI\n")
// 	// 	tc = &tls.Config{InsecureSkipVerify: true}
// 	// }k,
// 	if c.Protocol == "http" {
// 		fmt.Printf("Trying PLAIN-TEXT connection to ES...\n")
// 		tr := http.DefaultTransport.(*http.Transport).Clone()
// 		// client := &http.Client{Transport: tr}
// 		resp, b = ESDial(cr, c, t, mp, tr)

// 		if b == nil {
// 			c.Protocol = "https"
// 		} else {
// 			return resp, b
// 		}
// 	}

// 	if c.Protocol == "https" {
// 		fmt.Printf("Trying ENCRYPTED connection to ES\n")
// 		if len(cr.Cert) != 0 {
// 			if !CertCheck(c.Host, c.Port, c.SelfSigned, cr.Cert, true) {

// 				return nil, nil
// 			}
// 			if c.SelfSigned {
// 				// if !CertCheck(c.Host, c.Port, c.SelfSigned, cr.Cert, true) {

// 				// 	return nil, nil
// 				// }
// 				if IsPulic(cr.Cert) {
// 					fmt.Printf("\n%v\nThe server cert is publicly signed -- please change \"SelfSigned\" to \"false\" as currenlty this setting is incorrect\n%v", p.Stars, p.Stars)
// 					c.SelfSigned = false

// 				}
// 			}
// 			// fmt.Println(c.SelfSigned)
// 			// fmt.Println(IsPulic(cr.Cert))
// 			if !c.SelfSigned && !IsPulic(cr.Cert) {

// 				fmt.Printf("\n%v\nThe server cert is selfsigned -- please change \"SelfSigned\" to \"true\" as currenlty this setting is incorrect\n%v", p.Stars, p.Stars)
// 				c.SelfSigned = true

// 			}
// 			if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
// 				fmt.Println("Failed to append cert")
// 				tc = &tls.Config{RootCAs: pool}

// 			}
// 		} else {

// 			tr := http.DefaultTransport.(*http.Transport).Clone()
// 			resp, b = ESDial(cr, c, t, mp, tr)
// 			// if resp.StatusCode == 0 {

// 			// 	return resp, b
// 			// } else {
// 			fmt.Printf("\"es-certs\" doesn't have the expected certificate (or the secret doesn't exist at all)... ")
// 			if resp.StatusCode != 0 {
// 				fmt.Printf("Trying connectivity using Public PKI\nPlease create es-cert (per below) or change \"SelfSigned\" to \"false\"")
// 				CertCheck(c.Host, c.Port, c.SelfSigned, "", false)
// 				return resp, nil
// 			}
// 		}

// 	}
// 	tc = &tls.Config{InsecureSkipVerify: true}
// 	tc = &tls.Config{RootCAs: pool}

// 	tr := &http.Transport{TLSClientConfig: tc}
// 	return ESDial(cr, c, t, mp, tr)
// }

func ESDial(cr *collector.ES, c *collector.CPTelemetryStore, t string, mp bool, tr *http.Transport) (*http.Response, []byte) {

	var req *http.Request
	var resp *http.Response
	var path string
	var err error
	var b []byte

	// tc = &tls.Config{InsecureSkipVerify: true}
	// tr := &http.Transport{TLSClientConfig: tc}

	client := &http.Client{Transport: tr}

	path = c.Protocol + "://" + c.Host + ":" + c.Port

	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	req.SetBasicAuth(cr.Username, cr.Password)
	// req.Header.Set("name", "value")
	req.Header.Set("tsb-route-target", "elasticsearch")
	req.Header.Set("x-tetrate-token", t)

	// fmt.Println(req)

	resp, err = client.Do(req)

	if err != nil {
		resp = new(http.Response)
		resp.Status = err.Error()
		b = nil

	} else {
		b, err = io.ReadAll(resp.Body)

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return resp, b
}

// func MPCheck(cr *collector.ES, c *collector.CPTelemetryStore) (*http.Response, []byte) {
// 	var req *http.Request
// 	var resp *http.Response
// 	var client *http.Client
// 	var path string
// 	var err error
// 	var b []byte

// 	pool := x509.NewCertPool()

// 	path = c.Protocol + "://" + c.Host + ":" + c.Port

// 	// p := CustomPrint()
// 	if c.Protocol == "http" {
// 		fmt.Printf("Establishing PLAIN connection per CP Manifest settings\n")
// 		tr := http.DefaultTransport.(*http.Transport).Clone()
// 		client = &http.Client{Transport: tr}
// 	} else {
// 		fmt.Printf("Establishing SECURE connection per CP Manifest settings\n")
// 		if c.SelfSigned {

// 			if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
// 				fmt.Println("Failed to append cert")
// 			}

// 		}
// 		tc := &tls.Config{RootCAs: pool}
// 		tr := &http.Transport{TLSClientConfig: tc}
// 		client = &http.Client{Transport: tr}

// 	}
// 	req, err = http.NewRequest("GET", path, nil)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	req.SetBasicAuth(cr.Username, cr.Password)

// 	resp, err = client.Do(req)

// 	if err != nil {

// 		resp = new(http.Response)
// 		resp.Status = err.Error()
// 	} else {
// 		b, err = io.ReadAll(resp.Body)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}

// 	}
// 	return resp, b

// }

func VersionCheck(b []byte, v string) (bool, string) {
	// b, err := io.ReadAll(r.Body)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	data := ESResponse{}

	// err will mean there is an unexpected response from ES Server
	err := json.Unmarshal([]byte(b), &data)
	if err != nil {
		fmt.Println(err.Error())
		return false, "0"
	}

	if data.Version.Number[0:1] == v {
		return true, data.Version.Number[0:1]
	} else {
		return false, data.Version.Number[0:1]
	}

}

type ESResponse struct {

	// Version ESVersion
	Version struct {
		Number string
	}
}
