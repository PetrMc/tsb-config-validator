package validator

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func Checklist(cred *collector.ES, conn *collector.CPTelemetryStore) {

	var req *http.Request
	var resp *http.Response
	var client *http.Client
	var path string
	var err error
	pool := x509.NewCertPool()
	path = conn.Protocol + "://" + conn.Host + ":" + conn.Port

	if conn.SelfSigned {
		fmt.Println("selfsigned")

		if ok := pool.AppendCertsFromPEM([]byte(cred.Cert)); !ok {
			fmt.Println("Failed to append cert")
		}
	}
	tc := &tls.Config{RootCAs: pool}
	tr := &http.Transport{TLSClientConfig: tc}
	client = &http.Client{Transport: tr}

	// req, err = http.NewRequest("GET", path, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// This one line implements the authentication required for the task.

	// resp, err = client.Do(req)
	// if err != nil {
	// 	fmt.Println(err)

	// }

	// fmt.Println("Response status:", resp.Status)

	// } else {
	// 	path = "http://" + conn.Host + ":" + conn.Port
	// 	fmt.Println(path)
	// }

	req, err = http.NewRequest("GET", path, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.SetBasicAuth(cred.Username, cred.Password)
	// resp, err = client.Do(req)
	resp, err = client.Do(req)
	if err != nil {
		// println(err)
		return

	}

	fmt.Println("Response status:", resp.Status)

	// url := make([]string, 2)
	// url[0] = "https://" + conn.Host + ":" + conn.Port
	// fmt.Println(url[0])

	// req, err := http.NewRequest("GET", url[0], nil)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// This one line implements the authentication required for the task.
	// req.SetBasicAuth(cred.Username, cred.Password)

	// Make request and show output.

	// fmt.Println("Response status:", resp.Status)

	// scanner := bufio.NewScanner(resp.Body)
	// for i := 0; scanner.Scan() && i < 5; i++ {
	// 	fmt.Println(scanner.Text())
	// }
	// if err := scanner.Err(); err != nil {
	// 	panic(err)
	// 	fmt.Println(err)
	// 	return
	// }
}
