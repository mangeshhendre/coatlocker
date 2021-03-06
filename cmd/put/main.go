package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/drhayt/jwtclient"
)

func main() {

	var (
		insecure = flag.Bool("insecure", len(os.Getenv("COATLOCKER_INSECURE")) != 0, "COATLOCKER_INSECURE: Insecure, dont validate certificates")
		username = flag.String("user", os.Getenv("COATLOCKER_USERNAME"), "COATLOCKER_USERNAME: What username to use")
		password = flag.String("pass", os.Getenv("COATLOCKER_PASSWORD"), "COATLOCKER_PASSWORD: What password to use")
		authurl  = flag.String("authurl", os.Getenv("COATLOCKER_AUTHURL"), "COATLOCKER_AUTHURL: Url of authentication service")
		svcurl   = flag.String("svcurl", os.Getenv("COATLOCKER_SVCURL"), "COATLOCKER_SVCURL: Url of coatlocker")
		file     = flag.String("file", "", "file to upload")
		key      = flag.String("key", "", "upload key")
	)

	flag.Parse()

	if len(*file) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if len(*key) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if len(*svcurl) == 0 {
		flag.Usage()
		os.Exit(3)
	}

	if len(*authurl) == 0 {
		flag.Usage()
		os.Exit(4)
	}

	if len(*username) == 0 {
		flag.Usage()
		os.Exit(5)
	}

	if len(*password) == 0 {
		flag.Usage()
		os.Exit(6)
	}

	token, err := jwtclient.Authenticate(*insecure, *authurl, *username, *password)
	if err != nil {
		log.Fatalln("Error authenticating: ", err)
	}

	err = PutFileWithAuth(*insecure, *svcurl, *file, *key, token)
	if err != nil {
		log.Fatalf("Error processing request: %s", err)
		os.Exit(3)
	}
}

func PutFileWithAuth(insecure bool, url, file, key, token string) error {

	fileContents, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("unable to open file: %s", err)
	}

	defer fileContents.Close()

	// Setup the request
	request, err := http.NewRequest("PUT", url+"/"+key, fileContents)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

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
		return fmt.Errorf("unable to make request: %s", err)
	}

	defer webResponse.Body.Close()

	switch webResponse.StatusCode {
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("key already exists")
	case http.StatusInternalServerError:
		return fmt.Errorf("Server unavailable")
	case http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("Unexpected code received: %s", webResponse.Status)
	}

}

func usage() {
	fmt.Printf("Dont be a sucka man\n")
}
