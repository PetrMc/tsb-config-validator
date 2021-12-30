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

func BruteForce(cr *collector.ES, c collector.CPTelemetryStore, prt []string) {

	var r *http.Response
	var b []byte
	tconn := c

	p := CustomPrint()

	// prt := []string{"http", "https"}
	crt := []bool{true, false}

	for _, i := range prt {
		for _, ss := range crt {

			if i != c.Protocol || ss != c.SelfSigned {

				tconn.Protocol = i
				tconn.SelfSigned = ss
				fmt.Printf("\n%v\nTrying the following combination:\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", p.Stars, tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
				r, b = MPCheck(cr, &tconn)
				fmt.Printf("Response status:%v\n", r.Status)
			}
			if Codes(c, r, b) {
				return
			}
			tconn = c

		}
	}
}

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

func CheckFrontEnvoy(cr *collector.ES, c *collector.CPTelemetryStore, t string) (*http.Response, []byte) {
	var req *http.Request
	var resp *http.Response
	var client *http.Client
	var tc *tls.Config
	var path string
	var err error

	pool := x509.NewCertPool()

	path = "https://" + c.Host + ":" + c.Port

	fmt.Printf("\nEstablishing connection...\n")
	if len(cr.Cert) != 0 {
		if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
			fmt.Println("Failed to append cert")
			tc = &tls.Config{RootCAs: pool}
		}
	} else {
		fmt.Printf("\"es-certs\" doesn't have the expected certificate (or the secret doesn't exist at all)")
		tc = &tls.Config{InsecureSkipVerify: true}
	}

	tr := &http.Transport{TLSClientConfig: tc}

	client = &http.Client{Transport: tr}

	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.SetBasicAuth(cr.Username, cr.Password)
	// req.Header.Set("name", "value")
	req.Header.Set("tsb-route-target", "elasticsearch")
	req.Header.Set("x-tetrate-token", t)

	// fmt.Println(t, req.Header)

	resp, err = client.Do(req)

	if err != nil {
		resp = new(http.Response)
		resp.Status = err.Error()
	}
	b, err := io.ReadAll(resp.Body)
	// fmt.Printf("\nCbbbbb)\n")
	// fmt.Println(string(b[:]), resp.Status, resp.Body, b)
	if err != nil {
		fmt.Println(err)
	}

	return resp, b
}

func MPCheck(cr *collector.ES, c *collector.CPTelemetryStore) (*http.Response, []byte) {
	var req *http.Request
	var resp *http.Response
	var client *http.Client
	var path string
	var err error
	var b []byte

	pool := x509.NewCertPool()

	path = c.Protocol + "://" + c.Host + ":" + c.Port

	// p := CustomPrint()
	if c.Protocol == "http" {
		fmt.Printf("Establishing PLAIN connection per CP Manifest settings\n")
		tr := http.DefaultTransport.(*http.Transport).Clone()
		client = &http.Client{Transport: tr}
	} else {
		fmt.Printf("Establishing SECURE connection per CP Manifest settings\n")
		if c.SelfSigned {

			if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
				fmt.Println("Failed to append cert")
			}

		}
		tc := &tls.Config{RootCAs: pool}
		tr := &http.Transport{TLSClientConfig: tc}
		client = &http.Client{Transport: tr}

	}
	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.SetBasicAuth(cr.Username, cr.Password)

	resp, err = client.Do(req)

	if err != nil {

		resp = new(http.Response)
		resp.Status = err.Error()
	} else {
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
		}

	}
	return resp, b

}

func VersionCheck(b []byte, v string) (bool, string) {
	// b, err := io.ReadAll(r.Body)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	data := ESResponse{}
	err := json.Unmarshal([]byte(b), &data)
	if err != nil {
		fmt.Println(err.Error())
		return false, "0"
	}

	if data.Version.Number[0:1] == v {
		return true, data.Version.Number[0:1]
	} else {
		// p := CustomPrint()
		// fmt.Printf("\n%v\nElastic Search Version mismatch:\n Version specified in CP is %v\nWhile ES instance returns: %v\n%v", p.Stars, v, data.Version.Number[0:1], p.Stars)
		return false, data.Version.Number[0:1]
	}

}

type ESResponse struct {

	// Version ESVersion
	Version struct {
		Number string
	}
}
