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
	"github.com/PetrMc/tsb-config-validator/src/output"
)

func Checklist(cred *collector.ES, conn collector.CPTelemetryStore, tokens collector.TSBTokens, fe bool) {

	if fe {
		Same(cred, &conn, &tokens)
	} else {
		Diff(cred, conn)
	}
}

func Diff(cred *collector.ES, conn collector.CPTelemetryStore) {

	p := output.CustomPrint()

	fmt.Printf("\nChecking direct connection between CP and MP\n")

	r := CheckMP(cred, &conn)

	switch r.StatusCode {

	case 200:

		m, v := VersionCheck(r, conn.Version)
		if m {
			fmt.Printf("\n%v\nResponse status:%v\n%vThe settings seem to be working as it is - no additional checks are done\n", p.Stars, r.Status, p.Indent)
		} else {
			fmt.Printf("\n%v\nResponse status:%v\n", p.Stars, r.Status)
			fmt.Printf("\n%v\nHoever Elastic Search Version mismatch is detected:\nVersion specified in CP is %v\nWhile ES instance returns: %v\n%v", p.Indent, conn.Version, v, p.Stars)
		}
		// os.Exit(0)

	case 401:

		fmt.Printf("\nReceived HTTP Code: %v, which means credentials are not correctly specified in \"elastic-credentials\" secret in \"istio-system\"\n", r.StatusCode)
		PasswdCheck(cred)

	default:

		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - the series of tests will try different combination of setting to get to the bottom of the problem\n", p.Stars, r.Status, p.Indent)
		BruteForce(cred, &conn)

	}
}

func BruteForce(cr *collector.ES, c *collector.CPTelemetryStore) {

	var r *http.Response
	tconn := c

	p := output.CustomPrint()

	prt := []string{"http", "https"}
	crt := []bool{true, false}

	for _, i := range prt {
		for _, ss := range crt {

			if i != c.Protocol || ss != c.SelfSigned {

				tconn.Protocol = i
				tconn.SelfSigned = ss
				fmt.Printf("\n%v\nTrying the following combination:\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", p.Stars, tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
				r = CheckMP(cr, tconn)
				fmt.Printf("Response status:%v\n", r.Status)
			}
			if r.StatusCode == 200 {

				m, v := VersionCheck(r, tconn.Version)
				if v != "0" {
					fmt.Printf("\n%v\nBINGO !!!\nWe got the correct settings: \n", p.Stars)
					fmt.Printf("\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
					if !m {
						fmt.Printf("Additionally - please make sure ES Version is set to %v\n%v", v, p.Stars)
						// os.Exit(0)
						return
					}
				} else {
					fmt.Printf("\n%v\nNot so right...\nWe got the settings that produce correct code but ES is not responding with correct body: \n", p.Stars)
				}

			}
			tconn = c

		}
	}
}

func Same(cred *collector.ES, conn *collector.CPTelemetryStore, t *collector.TSBTokens) {

	p := output.CustomPrint()

	fmt.Printf("\nChecking connection between CP and FrontEnvoy (running in MP)\n")
	r := CheckFrontEnvoy(cred, conn, t.Zipkint)

	fmt.Printf("\n%v\nWhile the connection is established succesfully the page returned is not what is expected from ES.\nContinuing.... \n", p.Stars)

	fmt.Println(r.StatusCode)

}

func PasswdCheck(cr *collector.ES) {

	p := output.CustomPrint()

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

func CheckFrontEnvoy(cr *collector.ES, c *collector.CPTelemetryStore, t string) *http.Response {
	var req *http.Request
	var resp *http.Response
	var client *http.Client
	var path string
	var err error

	pool := x509.NewCertPool()

	path = "https://" + c.Host + ":" + c.Port
	// path = "https://cx-jwt-token-tsb.cx.tetrate.info:8443/"
	// fmt.Println(t)
	// p := output.CustomPrint()
	fmt.Printf("Establishing SECURE connection per CP Manifest settings\n")
	if c.SelfSigned {

		if ok := pool.AppendCertsFromPEM([]byte(cr.Cert)); !ok {
			fmt.Println("Failed to append cert")
		}

	}
	tc := &tls.Config{RootCAs: pool}
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
		return resp
	}
	b, err := io.ReadAll(resp.Body)
	myString := string(b[:])
	fmt.Println(myString)
	return resp
}

func CheckMP(cr *collector.ES, c *collector.CPTelemetryStore) *http.Response {
	var req *http.Request
	var resp *http.Response
	var client *http.Client
	var path string
	var err error

	pool := x509.NewCertPool()

	path = c.Protocol + "://" + c.Host + ":" + c.Port

	// p := output.CustomPrint()
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
		// resp.Status = err + "Can't connect to the server"
		resp.Status = err.Error()
		return resp
	}

	return resp

}

func VersionCheck(r *http.Response, v string) (bool, string) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(string(b))
	data := ESResponse{}
	err = json.Unmarshal([]byte(b), &data)
	if err != nil {
		// fmt.Println(err.Error())
		fmt.Println(data)

		return false, "0"
	}

	if data.Version.Number[0:1] == v {
		return true, data.Version.Number[0:1]
	} else {
		// p := output.CustomPrint()
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

// type ESVersion struct {
// 	Number string
// }

// func srcSecure() {
// 	var req *http.Request
// 	var resp *http.Response
// 	var client *http.Client
// 	var path string
// 	var err error
// 	pool := x509.NewCertPool()
// 	path = conn.Protocol + "://" + conn.Host + ":" + conn.Port
// 	p := output.CustomPrint()

// 	fmt.Printf("Establishing secure connection per CP Manifest settings")

// 	if conn.SelfSigned {

// 		if ok := pool.AppendCertsFromPEM([]byte(cred.Cert)); !ok {
// 			fmt.Println("Failed to append cert")
// 		}
// 	}

// 	tc := &tls.Config{RootCAs: pool}
// 	tr := &http.Transport{TLSClientConfig: tc}
// 	client = &http.Client{Transport: tr}
// 	// req, err = http.NewRequest("GET", path, nil)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }

// 	// This one line implements the authentication required for the task.

// 	// resp, err = client.Do(req)
// 	// if err != nil {
// 	// 	fmt.Println(err)

// 	// }

// 	// fmt.Println("Response status:", resp.Status)

// 	// } else {
// 	// 	path = "http://" + conn.Host + ":" + conn.Port
// 	// 	fmt.Println(path)
// 	// }

// 	req, err = http.NewRequest("GET", path, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	req.SetBasicAuth(cred.Username, cred.Password)
// 	// resp, err = client.Do(req)
// 	resp, err = client.Do(req)
// 	if err != nil {
// 		// println(err)
// 		return

// 	}

// 	fmt.Printf("\n%v\nResponse status:%v\n", p.Stars, resp.Status)

// 	// url := make([]string, 2)
// 	// url[0] = "https://" + conn.Host + ":" + conn.Port
// 	// fmt.Println(url[0])

// 	// req, err := http.NewRequest("GET", url[0], nil)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }

// 	// This one line implements the authentication required for the task.
// 	// req.SetBasicAuth(cred.Username, cred.Password)

// 	// Make request and show output.

// 	// fmt.Println("Response status:", resp.Status)

// 	// scanner := bufio.NewScanner(resp.Body)
// 	// for i := 0; scanner.Scan() && i < 5; i++ {
// 	// 	fmt.Println(scanner.Text())
// 	// }
// 	// if err := scanner.Err(); err != nil {
// 	// 	panic(err)
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// }
