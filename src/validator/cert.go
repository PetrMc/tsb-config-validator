package validator

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"time"
)

func CertCheck(h string, p string) {

	str := h + ":" + p
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", str, conf)
	if err != nil {
		fmt.Printf(err.Error())
	}

	err = conn.VerifyHostname(h)
	if err != nil {
		fmt.Printf("Hostname doesn't match with certificate: " + err.Error())
	}
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	fmt.Printf("Issuer: %s\nExpiry: %v\n", conn.ConnectionState().PeerCertificates[0].Issuer, expiry.Format(time.RFC850))
	fmt.Println(conn.ConnectionState().PeerCertificates)
	var b bytes.Buffer
	for _, cert := range conn.ConnectionState().PeerCertificates {
		err := pem.Encode(&b, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
		if err != nil {
			fmt.Printf(err.Error())
		}
	}
	fmt.Printf(b.String())
}
