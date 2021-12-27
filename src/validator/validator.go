package validator

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PetrMc/tsb-config-validator/src/collector"
	"github.com/PetrMc/tsb-config-validator/src/output"
)

func Checklist(cred *collector.ES, conn collector.CPTelemetryStore) {

	r := CheckMP(cred, &conn)
	p := output.CustomPrint()
	if r.StatusCode == 200 {
		m, v := VersionCheck(r, conn.Version)
		if m {
			fmt.Printf("\n%v\nResponse status:%v\n%vThe settings seem to be working as it is - no additional checks are done\n", p.Stars, r.Status, p.Indent)
		} else {
			fmt.Printf("\n%v\nResponse status:%v\n", p.Stars, r.Status)

			fmt.Printf("\n%v\nHoever Elastic Search Version mismatch is detected:\nVersion specified in CP is %v\nWhile ES instance returns: %v\n%v", p.Indent, conn.Version, v, p.Stars)
		}
		// os.Exit(0)

	} else {
		fmt.Printf("\n%v\nResponse status:%v\n%v\nThe settings are not working - the series of tests will try different combination of setting to get to the bottom of the problem\n", p.Stars, r.Status, p.Indent)
	}

	tconn := conn

	prt := []string{"http", "https"}
	crt := []bool{true, false}

combine:
	for _, i := range prt {
		for _, c := range crt {

			if i != conn.Protocol || c != conn.SelfSigned {

				tconn.Protocol = i
				tconn.SelfSigned = c
				fmt.Printf("\n%v\nTrying the following combination:\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", p.Stars, tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
				r = CheckMP(cred, &tconn)
				fmt.Printf("Response status:%v\n", r.Status)
			}
			if r.StatusCode == 200 {
				fmt.Printf("\n%v\nBINGO !!!\nWe got the correct settings: \n", p.Stars)
				fmt.Printf("\nHost - %v | Port - %v | Protocol - %v | Selfsigned - %v \n", tconn.Host, tconn.Port, tconn.Protocol, tconn.SelfSigned)
				m, v := VersionCheck(r, tconn.Version)
				if m {
					fmt.Printf("Additionally - please make sure ES Version is set to %v\n%v", v, p.Stars)
				}
				// os.Exit(0)
				break combine
			}
			tconn = conn

		}
	}

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
	json.Unmarshal([]byte(b), &data)

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
