package validator

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/PetrMc/tsb-config-validator/src/collector"
)

func Checklist(cred collector.ES, conn collector.CPTelemetryStore) {
	url := make([]string, 2)
	url[0] = "https://" + conn.Host + ":" + conn.Port
	fmt.Println(url[0])

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(cred.Cert)); !ok {
		fmt.Println("Failed to append cert")
	}
	tc := &tls.Config{RootCAs: pool}
	tr := &http.Transport{TLSClientConfig: tc}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url[0], nil)
	if err != nil {
		fmt.Println(err)
	}

	// This one line implements the authentication required for the task.
	req.SetBasicAuth(cred.Username, cred.Password)

	// Make request and show output.
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Response status:", resp.Status)

	// req.SetBasicAuth(cred.Username, cred.Password)

	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM([]byte(cred.Cert))

	// // Setup HTTPS client
	// tlsConfig := &tls.Config{
	// 	RootCAs: caCertPool,
	// }

	// transport := &http.Transport{TLSClientConfig: tlsConfig}
	// client := &http.Client{Transport: transport}

	// resp, err := client.Get(url[0])
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer resp.Body.Close()

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
