package jwtclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Authenticate is a jwt wrapper that returns the JWT token to be used on subsequent calls.
func Authenticate(insecure bool, url, username, password string) (token string, err error) {

	// Setup the request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Set our auth parameters.
	request.SetBasicAuth(username, password)

	// Get a config to override potential SSL stuff.
	tlsConfig := &tls.Config{}

	// If we are insecure, then so be it.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Create the transport
	tr := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a client using that transport.
	client := &http.Client{Transport: tr}

	// Make the request
	webResponse, err := client.Do(request)
	if err != nil {
		return
	}

	defer webResponse.Body.Close()
	buffer := &bytes.Buffer{}
	io.Copy(buffer, webResponse.Body)
	token = buffer.String()
	return
}

// RetrieveCertificate is a wrapper to retrieve a certificate from a remote server.
func RetrieveCertificate(insecure bool, url string) (certificate string, err error) {

	// Setup the request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Get a config to override potential SSL stuff.
	tlsConfig := &tls.Config{}

	// If we are insecure, then so be it.
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// Create the transport
	tr := &http.Transport{TLSClientConfig: tlsConfig}

	// Create a client using that transport.
	client := &http.Client{Transport: tr}

	// Make the request
	webResponse, err := client.Do(request)
	if err != nil {
		return
	}

	defer webResponse.Body.Close()
	buffer := &bytes.Buffer{}
	io.Copy(buffer, webResponse.Body)
	certificate = buffer.String()
	certificate = strings.Replace(certificate, "-----BEGIN CERTIFICATE-----", fmt.Sprintf("-----BEGIN CERTIFICATE-----\n"), -1)
	certificate = strings.Replace(certificate, "-----END CERTIFICATE-----", fmt.Sprintf("\n-----END CERTIFICATE-----\n"), -1)
	return
}

// RetrieveCertificate is a wrapper to retrieve a certificate from a remote server.
func KeyFuncClosure(insecure bool, url string) (jwt.Keyfunc, error) {

	certificate, err := RetrieveCertificate(insecure, url)
	if err != nil {
		return nil, err
	}

	// Decode the certificate
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return nil, fmt.Errorf("unable to decode pem encoded certificate")
	}

	pub, err := x509.ParseCertificate(block.Bytes)
	// pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: " + err.Error())
	}

	return func(*jwt.Token) (interface{}, error) { return pub.PublicKey, nil }, nil
}
