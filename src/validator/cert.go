package validator

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func CertCheck(h string, po string, ss bool, k8scert string, x bool) bool {

	var conf *tls.Config
	// pool := x509.NewCertPool()

	str := h + ":" + po
	// if ok := pool.AppendCertsFromPEM([]byte(k8scert)); !ok {
	// 	fmt.Println("Failed to append cert")
	// }
	conf = &tls.Config{
		InsecureSkipVerify: true,
		// InsecureSkipVerify: false,
		// RootCAs:            pool,
	}

	conn, err := tls.Dial("tcp", str, conf)
	if err != nil {
		fmt.Printf(err.Error())
	}

	if !ss {
		err = conn.VerifyHostname(h)
		if err != nil {
			fmt.Printf("Hostname doesn't match with certificate: " + err.Error())
		}
	}

	// fmt.Println(conn.ConnectionState().PeerCertificates)
	var cacert bytes.Buffer
	// for _, cert := range conn.ConnectionState().PeerCertificates {
	// err := pem.Encode(&b, &pem.Block{
	// 	Type:  "CERTIFICATE",
	// 	Bytes: cert.Raw,
	// })
	// if err != nil {
	// 	fmt.Printf(err.Error())
	// }
	// }
	// lastcert := conn.ConnectionState().PeerCertificates[len(conn.ConnectionState().PeerCertificates)-1]
	err = pem.Encode(&cacert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: conn.ConnectionState().PeerCertificates[len(conn.ConnectionState().PeerCertificates)-1].Raw,
	})
	// err := pem.Encode(&b, &pem.Block{
	// 	Type:  "CERTIFICATE",
	// 	Bytes: cert.Raw,
	// })
	if err != nil {
		fmt.Printf(err.Error())
	}
	// fmt.Println(k8scert)
	return IsMatch(cacert.String(), k8scert, x)
}
func IsMatch(srv string, k8s string, x bool) bool {
	var m string
	var match bool
	var err error
	p := CustomPrint()
	// fmt.Println(srv, k8s)
	if srv != k8s {
		match = false
		if x {
			m = "\n" + p.Stars + "\nThe certificates don't match\n"
		} else {
			// m = "\nThe certificate in \"es-cert\" secret (\"istio-system\" namespace) doesn't exist\n"
			m = ""
		}
		fmt.Println(m)
		fn := "/tmp/ca.crt"
		CASave(srv, fn)
		if err != nil {
			cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=ca.crt\n"
			fmt.Printf("\nThe below cert needs to be added to k8s cluster\nPlease copy to ca.crt file and then run the folloing command\n%v\n%v\n", cmd, srv)
		} else {
			cmd := p.Indent + "kubectl -n istio-system get secret es-certs -o yaml > /tmp/es-certs-backup.yaml\n" + p.Indent + "kubectl -n istio-system delete secret es-certs\n" + p.Indent + "kubectl -n istio-system create secret generic es-certs  --from-file=ca.crt=" + fn + "\n"
			fmt.Printf("\nPlease run the commands below to re-created the secret from to %v file and then run the folloing command\n%v\n", fn, cmd)
		}
	} else {
		match = true
	}

	//  expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	//  fmt.Printf("Issuer: %s\nExpiry: %v\n", conn.ConnectionState().PeerCertificates[0].Issuer, expiry.Format(time.RFC850))
	return match
}

func IsPulic(k8scert string) bool {
	block, _ := pem.Decode([]byte(k8scert))
	p := false
	if block == nil {
		panic("failed to parse certificate PEM")

	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	if len(cert.AuthorityKeyId) != 0 {
		p = true

	}
	// fmt.Println(cert.AuthorityKeyId)
	return p
}

func CASave(cacert string, filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	} else {
		file.WriteString(cacert)
		fmt.Printf("\nThe certificate is saved to %v file in the current directory\n", filename)
	}
	file.Close()
	return err
}
