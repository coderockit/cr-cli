package crcli

import (
	"strconv"
	"strings"

	"github.com/juju/loggo"
	"gopkg.in/resty.v1"
)

// The CodeRockIt pin type
type Pin struct {
	PinVerb string
	// The crpin url is of the form -- crpin://[host]:[port]/[cr-group-name]/[cr-name]/[cr-version]/[cr-hash-verifier]
	Host      string
	Port      int
	GroupName string
	Name      string
	Version   string
	Hash      string
	Verified  bool
}

// The CodeRockIt pinmap type -- map file paths to their contained pins
type Pinmap map[string][]Pin

func NewPin(verb string, pinStr string) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	//logger.Debugf("Creating new pin: %s", pinStr)

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
	version := parts[3]
	hash := parts[4]

	newPin := Pin{
		PinVerb: verb, Host: host, Port: port,
		GroupName: groupName, Name: name,
		Version: version, Hash: hash, Verified: false,
	}
	//logger.Debugf("returning newPin: %s", newPin)
	return newPin
}

func parsePin(verb string, pinStr string) []string {
	//logger := loggo.GetLogger("coderockit.cli.crcli")
	var parts []string

	// [verb] crpin://cooderocket.io/asdf/asdfasdfasdf/v1.1.1/
	beginIndex := strings.Index(pinStr, verb+" crpin:")
	if beginIndex != -1 {
		realPin := strings.Split(pinStr[beginIndex:len(pinStr)], " ")
		//logger.Debugf("realPin 0: %s", realPin[0])
		parts = strings.Split(realPin[1][8:len(realPin[1])], "/")
	}

	//logger.Debugf("parsePin parts is: %s", parts)
	return parts
}

func getResourceURL(pin Pin) string {
	port := ":" + strconv.Itoa(pin.Port)
	if pin.Port == 80 || pin.Port == 443 {
		port = ""
	}
	return ConfString("protocol", "https") + "://" + pin.Host + port + "/" + ConfString("pinResource", "pin") + "/"
}

func getVerifyURL(pin Pin) string {
	return getResourceURL(pin) + pin.GroupName + "/" + pin.Name + "/" + pin.Version + "/" + pin.Hash
}

func verifyPin(pin Pin) {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(ConfString("apiAccessToken", "")).
		Get(getVerifyURL(pin))

	if err == nil {
		logger.Debugf("resonse body: %s", resp.Body())
	} else {
		logger.Criticalf("Error: %s", err)
	}
}
