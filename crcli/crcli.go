package crcli

import (
	"crypto/tls"
	"encoding/json"
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
	GroupName     string
	Name          string
	HasParts      bool
	Version       string
	ApplyVersion  string
	Hash          string
	ParentVersion string
	IsPrivate     bool
	ApiMsg        string
}

func (p Pin) String() string {
	return fmt.Sprintf("%s %s/%s/%s/%s/%s", p.Verb, p.GroupName, p.Name, p.Version, p.Hash, p.ParentVersion)
}

func (p Pin) IsGet() bool {
	return strings.Contains(p.Verb, "GET")
}

func (p Pin) IsPut() bool {
	return strings.Contains(p.Verb, "PUT")
}

func (p Pin) ApiSuccess() bool {
	return strings.HasPrefix(p.ApiMsg, "Success")
}

// The CodeRockIt pinmap type -- map file paths to their contained pins
type Pinmap map[string][]Pin

func NewPin(verb string, pinUri string) Pin {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	pinUri = strings.TrimSpace(pinUri)

	logger.Debugf("%s :: Creating new pin: %s", verb, pinUri)

	parts := parsepinUri(verb, pinUri)
	logger.Debugf("parts is: %s", parts)

	groupName := parts[0]
	name := parts[1]
	hasParts := strings.Contains(name, "::")
	version := "NONE"
	if len(parts) >= 3 {
		version = parts[2]
	}
	parentVersion := "NONE"
	if len(parts) >= 4 {
		parentVersion = parts[3]
	}
	isPrivate := (strings.Index(verb, "PRIVATE") != -1)

	newPin := Pin{
		Verb: verb, IsPrivate: isPrivate,
		GroupName: groupName, Name: name, HasParts: hasParts,
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

func getVerifyPinURI(apiURL string, pin Pin) string {
	logger := loggo.GetLogger("coderockit.cli.crcli")
	logger.Debugf("Escaping pin version: %s", pin.Version)
	//logger.Debugf("Pin hash: %s", pin.Hash)
	//versionAndHashAndParent := ""

	if pin.IsGet() {
		return apiURL + "/pin/" + pin.GroupName + "/" + pin.Name + "/" +
			url.PathEscape(pin.Version)
	} else if pin.IsPut() {
		return apiURL + "/pin/" + pin.GroupName + "/" + pin.Name + "/" +
			url.PathEscape(pin.Version) + "/" + UrlEncodeBase64(pin.Hash) + "/" +
			pin.ParentVersion + "/" + strconv.FormatBool(pin.IsPrivate)
	}

	return ""
}

func verifyPin(pin Pin, pinContent string, sendContent bool) Pin {
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
		verifyURL := getVerifyPinURI(apiURL, pin)
		logger.Debugf("verifying %s pin with URL: %s", pin.Verb, verifyURL)

		var err error
		if pin.IsGet() {
			resp, err := resty.R().
				SetHeader("Accept", "application/json").
				SetAuthToken(GetApiAccessToken(tokIndex)).
				Get(verifyURL)

			// The response should contain at most the last three
			// versions of this pin, if empty then there was an error
			// trying to GET this pin - the ApiMsg explains why

			pin, err = handleVerifyPinResponse(pin, err, resp)
		} else if pin.IsPut() {
			contentToSend := pinContent
			if !sendContent {
				contentToSend = ""
			}
			resp, err := resty.R().
				SetHeader("Accept", "application/json").
				SetAuthToken(GetApiAccessToken(tokIndex)).
				SetHeader("Content-Type", "application/octet-stream").
				SetBody(contentToSend).
				Put(verifyURL)

			// The response should contain at most the last three
			// versions of this pin, if empty then there was an error
			// trying to PUT this pin - the ApiMsg explains why

			pin, err = handleVerifyPinResponse(pin, err, resp)
		}

		if err == nil {
			break
		}
	}

	if pin.IsPut() {
		WritePinContentToApply(pin, pinContent)
	}

	return pin
}

func handleVerifyPinResponse(pin Pin, err error, resp *resty.Response) (Pin, error) {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	var myerr error
	if err == nil {
		logger.Debugf("resonse status code: %d", resp.StatusCode())
		if resp.StatusCode() == 200 {
			respBody := resp.Body()
			var respObj interface{}
			err := json.Unmarshal(respBody, &respObj)
			if err == nil {
				respMap := respObj.(map[string]interface{})
				logger.Debugf("Reponse from verifying pin with CORRECT object: %s", respMap)
				pin.ApplyVersion = respMap["applyVersion"].(string)
				pin.ApiMsg = fmt.Sprintf("Success: %s", respMap["message"].(string))
				if pin.IsGet() {
					pin.Hash = respMap["hash"].(string)
					WritePinApiContentToCache(pin, respMap["content"].(string))
				}
			} else {
				logger.Debugf("Error INCORRECT json in response %s: %s", respBody, err)
				pin.ApiMsg = fmt.Sprintf("Fatal: could not parse response JSON: %s :: %s", respBody, err)
			}
		} else {
			respBody := resp.Body()
			var respObj interface{}
			err := json.Unmarshal(respBody, &respObj)
			if err == nil {
				respMap := respObj.(map[string]interface{})
				logger.Debugf("Reponse from verifying pin with CORRECT object: %s", respMap)
				pin.ApplyVersion = respMap["applyVersion"].(string)
				pin.ApiMsg = fmt.Sprintf("Fatal %d: verification failed with error: %s", resp.StatusCode(), respBody)
				myerr = errors.New(string(resp.StatusCode()))
			} else {
				logger.Debugf("Error INCORRECT json in response %s: %s", respBody, err)
				pin.ApiMsg = fmt.Sprintf("Fatal: could not parse response JSON: %s :: %s", respBody, err)
			}
			//respBody := string(resp.Body())
			//pin.ApiMsg = fmt.Sprintf("Fatal %d: verification failed with error: %s", resp.StatusCode(), respBody)
		}
	} else {
		myerr = err
		pin.ApiMsg = fmt.Sprintf("Error: %s", err)
		logger.Criticalf("Error: %s", err)
	}

	return pin, myerr
}

func GetMatchingVersions(requirement string, versions []string) []string {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	var matchingVersions []string
	for _, version := range versions {
		if version == requirement {
			matchingVersions = append(matchingVersions, version)
		}
	}

	if matchingVersions == nil {

		if ConfBool("apiAllowInsecure", false) {
			resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}

		apiURLs := ConfStringSlice("apiURLs", []string{"https://coderockit.io/api/v1"})
		for tokIndex, apiURL := range apiURLs {

			matchingVersionsURL := apiURL + "/matchingVersions/" + url.PathEscape(requirement) + "/" +
				url.PathEscape(strings.Join(versions, ","))

			logger.Debugf("Getting matching versions '%s' using URL: %s", versions, matchingVersionsURL)

			resp, err := resty.R().
				SetHeader("Accept", "application/json").
				SetAuthToken(GetApiAccessToken(tokIndex)).
				Get(matchingVersionsURL)

			if err == nil {
				if resp.StatusCode() == 200 {
					respBody := resp.Body()
					var respObj interface{}
					err := json.Unmarshal(respBody, &respObj)
					if err == nil {
						respMap := respObj.(map[string]interface{})
						respVersions := respMap["versions"].([]interface{})
						for _, respVer := range respVersions {
							matchingVersions = append(matchingVersions, fmt.Sprintf("%s", respVer))
						}

						logger.Debugf("Response from get matching versions CORRECT object: %s", respMap)

						if matchingVersions != nil {
							break
						}
					} else {
						logger.Debugf("Error INCORRECT json in response %s: %s", respBody, err)
						//pin.ApiMsg = fmt.Sprintf("Fatal: could not parse response JSON: %s :: %s", respBody, err)
					}
				} else {
					//myerr = errors.New(string(resp.StatusCode()))
					respBody := string(resp.Body())
					logger.Debugf("Fatal %d: matching versions failed with error: %s", resp.StatusCode(), respBody)
				}
			} else {
				logger.Criticalf("Error: %s", err)
			}
		}
	}

	return matchingVersions
}
