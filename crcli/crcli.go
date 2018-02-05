package crcli

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	"github.com/juju/loggo"
	"gopkg.in/resty.v1"
)

// The CodeRockIt pin type
type Pin struct {
	PinVerb string
	// The pin url is of the form -- pin://[host]:[port]/[cr-group-name]/[cr-name]/[cr-version]/[cr-hash-verifier]
	Host      string
	Port      int
	GroupName string
	Name      string
	Version   string
	Hash      string
	ErrorMsg  string
}

// The CodeRockIt pinmap type -- map file paths to their contained pins
type Pinmap map[string][]Pin

func NewPin(fullFilepath string, verb string, pinStr string) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	//logger.Debugf("%s :: Creating new pin: %s", verb, pinStr)

	parts := parsePin(verb, pinStr)
	//logger.Debugf("parts is: %s", parts)

	//verb := "GET"
	host := parts[0] //"coderockit.io"
	port := 80
	portIndex := strings.Index(host, ":")
	if portIndex != -1 {
		var err error
		port, err = strconv.Atoi(host[portIndex+1 : len(host)])
		if err != nil {
			port = 80
			logger.Criticalf("Your port is not a number: %s", pinStr)
		}
		host = host[0:portIndex]
	}
	groupName := parts[1]
	name := parts[2]
	version := ""
	if len(parts) >= 4 {
		version = parts[3]
	}
	hash := ""
	if len(parts) >= 5 {
		hash = parts[4]
	}

	newPin := Pin{
		PinVerb: verb, Host: host, Port: port,
		GroupName: groupName, Name: name,
		Version: version, Hash: hash,
		ErrorMsg: "No attempt to verify yet",
	}
	//logger.Debugf("returning newPin: %s", newPin)
	return newPin
}

func parsePin(verb string, pinStr string) []string {
	//logger := loggo.GetLogger("coderockit.cli.crcli")
	var parts []string

	// [verb] pin://cooderocket.io/asdf/asdfasdfasdf/v1.1.1/
	beginIndex := strings.Index(pinStr, verb+" pin:")
	if beginIndex != -1 {
		realPin := strings.Split(pinStr[beginIndex:len(pinStr)], " ")
		//logger.Debugf("realPin 0: %s", realPin[0])
		parts = strings.Split(realPin[1][6:len(realPin[1])], "/")
	}

	//logger.Debugf("%s :: %s :: parsePin parts is: %s", verb, pinStr, parts)
	return parts
}

func getResourceURL(pin Pin) string {
	port := ":" + strconv.Itoa(pin.Port)
	if pin.Port == 80 || pin.Port == 443 {
		port = ""
	}
	return ConfString("apiProtocol", "https") + "://" + pin.Host + port + "/" + ConfString("apiPinResource", "api/v1/pin") + "/"
}

func getVerifyURL(pin Pin) string {
	//logger := loggo.GetLogger("coderockit.cli.crcli")
	//logger.Debugf("Pin version: %s", pin.Version)
	//logger.Debugf("Pin hash: %s", pin.Hash)
	versionAndHash := ""
	if pin.Version != "" {
		versionAndHash += "/" + pin.Version
	}
	if pin.Hash != "" {
		versionAndHash += "/" + pin.Hash
	}
	return getResourceURL(pin) + pin.GroupName + "/" + pin.Name + versionAndHash
}

func verifyPin(pin Pin) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	//resty.SetProxy("http://127.0.0.1:8080")
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	verifyURL := getVerifyURL(pin)
	logger.Debugf("verifying pin with URL: %s", verifyURL)
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(GetApiAccessToken()).
		Get(verifyURL)

	if err == nil {
		if resp.StatusCode() == 200 {
			//pin.Verified = string(resp.Body())
			pin.ErrorMsg = ""
		} else {
			//logger.Debugf("resonse body: %s", resp.Body())
			//pin.Verified = false
			respBody := string(resp.Body())
			pin.ErrorMsg = fmt.Sprintf("Fatal: verification failed with error: %s", respBody)
		}
	} else {
		//pin.Verified = false
		pin.ErrorMsg = fmt.Sprintf("%s", err)
		logger.Criticalf("Error: %s", err)
	}

	return pin
}
