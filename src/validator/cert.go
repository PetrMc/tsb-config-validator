package validator

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// SRVCert function uses crypto libraries to reach ElasticSearch
// tries to retrieve and decode the certificate presented by the
// server. If successful, returns true and the certificate, otherwise
// returns false and empty string

func SRVCert(h string, po string) (bool, string) {

	// building default config
	var conf *tls.Config
	// buffer is used for the cert decoding step
	var cacert bytes.Buffer

	// str is used as the address fot TLS call
	str := h + ":" + po

	// the goal here is to receive any certificate for analysis
	// security is not a concern here. It's the operator's
	// responsibility to make sure the setup is secure.
	conf = &tls.Config{
		InsecureSkipVerify: true,
	}

	// making a simple TCP call to ES Address/port from the function input parameters
	conn, err := tls.Dial("tcp", str, conf)
	// if any errors - the function prints the error to the terminal, returns false
	// and empty string for certificate placeholder
	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("\nNo certificate check can be done against the server per the following: %v.", err.Error())
		return false, ""
	}

	// the function is interested in last certificate in the chain, as if exists it can be CA cert
	// that the users can be utilized by the caller to establish trusted connection
	lastcert := conn.ConnectionState().PeerCertificates[len(conn.ConnectionState().PeerCertificates)-1]

	// the certificate received by tls.Dial requires encoding
	err = pem.Encode(&cacert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: lastcert.Raw,
	})

	// if any errors - the function prints the error to the terminal, returns false
	// and empty string for certificate placeholder
	if err != nil {
		fmt.Printf(err.Error())
		return false, ""

	}

	// if the call to the server is successful - true (meaning there is a parsable cert presented
	// by the server) and the last cert in the chain as string are returned by the function
	return true, cacert.String()
}

// IsMatch function analyzes the certificate provided as srv variable
// looks if the certificate is CA. Compares it to certificate from k8s string
// and if the CA cert is found but not matching k8s - suggest user to update the secret
// that holds the CA cert (es-certs) in the Kubernetes cluster.

func IsMatch(srv string, k8s string) bool {

	// The result of the function will be "true" - the CA is found and matches K8s secret
	// or "false" in any other scenario

	var match bool

	// first check if CA is provided by the server, if not there is no point to compare the secrets
	if !IsCA(srv) {

		fmt.Printf("\nCan't obtain CA cert from the server\n")
		// if the certificate stored in k8s secret CA - it can be used for testing the connection
		if IsCA(k8s) {
			fmt.Printf("\nThe certificate stored in \"es-certs\" in \"istio-system\" is a CA cert will try to use it\n")
			// to satisfy the higher level logic - the assumption is done that - k8s secret is sufficient to proceed
			return true
		} else {
			// if the certificate stored in es-cert doesn't exist, corrupted or not CA - there is no reason to proceed with additional testing
			fmt.Printf("\nThe certificate stored in \"es-certs\" in \"istio-system\" is not a CA\nYou have to obtain it manually")
			return false
		}
	}

	// Checking if the server and kubernetes secrets match
	if srv != k8s {
		match = false
	} else {
		match = true
	}

	return match
}

// certparse function uses standard approach to get make the certificate readable
func certparse(uc string) *x509.Certificate {

	// decoded certificate placeholder
	block, _ := pem.Decode([]byte(uc))

	// panic is not often used in this code - probably needs to be reviewed
	if block == nil {
		panic("failed to parse certificate PEM")
	}

	// Parsing the certificate metadata
	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	return cert
}

// IsExpired function is to check if the cert is expired
// not sure where to call it (currently not called) as
// we only check selfsigned CA while some of public CAs also
// can expire - i.e. Lets Encrypt CA has expired in Oct 2021

func IsExpired(uc string) bool {
	exp := false
	cert := certparse(uc)

	et := cert.NotAfter
	ct := time.Now()
	if et.Before(ct) {
		exp = true
		fmt.Printf("\nThe certificate seems to expire - Not After states: %v while the current time is %v\n", et.Format(time.RFC850), ct.Format(time.RFC850))
	}

	return exp
}

// IsPublic confirms if the certificate is SelfSigned
// this is a very important check as it directly correlates
// with CP Settings and if set incorrectly will not work
// fortunattely there is a way to automatically confirm that

func IsPublic(uc string) bool {

	// setting the default value
	p := false

	// getting readable certificate metadata
	cert := certparse(uc)

	// the certificate is public when (a) it has AuthorityKeyId
	// and (b) the value is different from cert.SubjectKeyId
	if len(cert.AuthorityKeyId) != 0 && !bytes.Equal(cert.AuthorityKeyId, cert.SubjectKeyId) {
		p = true
	}

	return p
}

// IsCA function checks if the provided certificate has
// CA field set to TRUE value
func IsCA(k8scert string) bool {

	// setting the default value
	p := false

	// getting readable certificate metadata
	cert := certparse(k8scert)

	// cheking the single certificate field value
	if cert.IsCA {
		p = true

	}
	// fmt.Println(cert.AuthorityKeyId)
	return p
}

// CASave writes the string value of the certificate to a file
// also prints out the name of the file if successful
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
