package crcli

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/juju/loggo"
	"gopkg.in/resty.v1"
)

// The CodeRockIt pin type
type Pin struct {
	Verb          string
	IsPrivate     bool
	GroupName     string
	Name          string
	Version       string
	ParentVersion string
	Hash          string
	ApiMsg        string
}

// The CodeRockIt pinmap type -- map file paths to their contained pins
type Pinmap map[string][]Pin

func NewPin(verb string, pinUri string) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	pinUri = strings.TrimSpace(pinUri)

	logger.Debugf("%s :: Creating new pin: %s", verb, pinUri)

	parts := parsepinUri(verb, pinUri)
	logger.Debugf("parts is: %s", parts)

	//verb := "GET"
	//	host := parts[0] //"coderockit.io"
	//	port := 80
	//	portIndex := strings.Index(host, ":")
	//	if portIndex != -1 {
	//		var err error
	//		port, err = strconv.Atoi(host[portIndex+1 : len(host)])
	//		if err != nil {
	//			port = 80
	//			logger.Criticalf("Your port is not a number: %s", pinUri)
	//		}
	//		host = host[0:portIndex]
	//	}

	groupName := parts[0]
	name := parts[1]
	version := "NONE"
	if len(parts) >= 3 {
		version = parts[2]
	}
	//	hash := ""
	//	if len(parts) >= 5 {
	//		hash = parts[4]
	//	}
	parentVersion := "NONE"
	if len(parts) >= 4 {
		parentVersion = parts[3]
	}
	isPrivate := (strings.Index(verb, "PRIVATE") != -1)

	newPin := Pin{
		Verb: verb, IsPrivate: isPrivate,
		GroupName: groupName, Name: name,
		Version: version, ParentVersion: parentVersion,
		Hash: "NONE", ApiMsg: "No attempt to verify yet",
	}
	logger.Debugf("returning newPin: %s", newPin)
	return newPin
}

func parsepinUri(verb string, pinUri string) []string {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	var parts []string

	// The pin url is of the form
	// [GET|PUT|PUTPRIVATE] /pin/[group-name]/[name]/[version]/[forked-pin-version]
	beginIndex := strings.Index(pinUri, verb+" /pin/")
	if beginIndex != -1 {
		realPin := strings.Split(pinUri[beginIndex:len(pinUri)], " ")
		logger.Debugf("realPin 0: %s", realPin[0])
		parts = strings.Split(realPin[1][5:len(realPin[1])], "/")
	}

	logger.Debugf("%s :: %s :: parsePin parts is: %s", verb, pinUri, parts)
	return parts
}

//func getResourceURI(pin Pin) string {
//	port := ":" + strconv.Itoa(pin.Port)
//	if pin.Port == 80 || pin.Port == 443 {
//		port = ""
//	}
//	return ConfString("apiProtocol", "https") + "://" + pin.Host + port + "/" + ConfString("apiPinResource", "api/v1/pin") + "/"
//}

func getVerifyURI(apiURL string, pin Pin) string {
	//logger := loggo.GetLogger("coderockit.cli.crcli")
	//logger.Debugf("Pin version: %s", pin.Version)
	//logger.Debugf("Pin hash: %s", pin.Hash)
	//versionAndHashAndParent := ""

	if strings.Index(pin.Verb, "GET") == 0 {
		//if pin.Version != "" {
		//	versionAndHashAndParent += "/" + pin.Version
		//} else {
		//	versionAndHashAndParent += "/NONE"
		//}
		return apiURL + "/" + pin.GroupName + "/" + pin.Name + "/" +
			url.PathEscape(pin.Version)
	} else if strings.Index(pin.Verb, "PUT") == 0 {
		//if pin.Version != "" {
		//	versionAndHashAndParent += "/" + pin.Version
		//} else {
		//	versionAndHashAndParent += "/NONE"
		//}
		//if pin.Hash != "" {
		//	versionAndHashAndParent += "/" + pin.Hash
		//} else {
		//	versionAndHashAndParent += "/NONE"
		//}
		//if pin.ParentVersion != "" {
		//	versionAndHashAndParent += "/" + pin.ParentVersion
		//} else {
		//	versionAndHashAndParent += "/NONE"
		//}
		return apiURL + "/" + pin.GroupName + "/" + pin.Name + "/" +
			url.PathEscape(pin.Version) + "/" + UrlEncodeBase64(pin.Hash) + "/" +
			pin.ParentVersion + "/" + strconv.FormatBool(pin.IsPrivate)
	}

	return ""
}

func verifyPin(pin Pin, pinContent string) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	//resty.SetProxy("http://127.0.0.1:8080")
	//logger.Debugf("The apiAllowInsecure flag is: %b", ConfBool("apiAllowInsecure", false))
	if ConfBool("apiAllowInsecure", false) {
		resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	// calculate hash of pinContent
	pin.Hash = Hash(pinContent)

	apiURLs := ConfStringSlice("apiURLs", []string{"https://coderockit.io/api/v1"})
	for tokIndex, apiURL := range apiURLs {
		verifyURL := getVerifyURI(apiURL, pin)
		logger.Debugf("verifying %s pin with URL: %s", pin.Verb, verifyURL)

		var err error
		if strings.Index(pin.Verb, "GET") == 0 {
			resp, err := resty.R().
				SetHeader("Accept", "application/json").
				SetAuthToken(GetApiAccessToken(tokIndex)).
				Get(verifyURL)

			// The response should contain at most the last three
			// versions of this pin, if empty then there was an error
			// trying to GET this pin - the ApiMsg explains why

			pin, err = handleResponse(pin, err, resp, true)
		} else if strings.Index(pin.Verb, "PUT") == 0 {
			resp, err := resty.R().
				SetHeader("Accept", "application/json").
				SetAuthToken(GetApiAccessToken(tokIndex)).
				SetBody("").
				Put(verifyURL)

			// The response should contain at most the last three
			// versions of this pin, if empty then there was an error
			// trying to PUT this pin - the ApiMsg explains why

			pin, err = handleResponse(pin, err, resp, false)
		}

		if err == nil {
			break
		}
	}

	return pin
}

func handleResponse(pin Pin, err error, resp *resty.Response, isGet bool) (Pin, error) {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	var err1 error
	if err == nil {
		logger.Debugf("resonse status code: %d", resp.StatusCode())
		if resp.StatusCode() == 200 {
			//pin.Verified = string(resp.Body())
			//err1 := nil
			respBody := string(resp.Body())
			pin.ApiMsg = fmt.Sprintf("Success: %s", respBody)
		} else {
			//pin.Verified = false
			err1 = errors.New(string(resp.StatusCode()))
			respBody := string(resp.Body())
			pin.ApiMsg = fmt.Sprintf("Fatal %d: verification failed with error: %s", resp.StatusCode(), respBody)
		}
	} else {
		//pin.Verified = false
		err1 = err
		pin.ApiMsg = fmt.Sprintf("Error: %s", err)
		logger.Criticalf("Error: %s", err)
	}

	return pin, err1
}
