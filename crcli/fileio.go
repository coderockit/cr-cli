package crcli

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/juju/loggo"
)

func SavePinCache(pinCache Pinmap) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	//logger.Debugf("Saving pin cache: %s", pinCache)

	pinCachePath := GetWorkDirectory() + "/pincache.json"
	jsonString, err := json.Marshal(pinCache)
	if err == nil {
		//logger.Debugf("JSON to save: %s", jsonString)
		err := ioutil.WriteFile(pinCachePath, jsonString, 0644)
		if err != nil {
			logger.Debugf("Error: %s", err)
		} else {
			logger.Debugf("Wrote pin cache to: %s", pinCachePath)
		}
	} else {
		logger.Debugf("Error: %s", err)
	}
}

func ReadInPinCache(pinCache Pinmap) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	logger.Debugf("Reading pin cache")
}

func ProcessPath(addPath string, pinCache Pinmap) Pinmap {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	abs, err := filepath.Abs(addPath)
	//logger.Debugf("Processing path: %s", abs)
	if err == nil {
		fi, err := os.Stat(abs)
		if err == nil {
			switch mode := fi.Mode(); {
			case mode.IsDir():
				// do directory stuff
				//fmt.Println("directory")
				allFiles, err := ioutil.ReadDir(abs)
				if err == nil {
					for _, nextFile := range allFiles {
						pinCache = ProcessPath(abs+"/"+nextFile.Name(), pinCache)
					}
				}
			case mode.IsRegular():
				// do file stuff
				//fmt.Println("file")
				newPins := GetPins(addPath)
				//logger.Debugf("!!!newPins is: %s", newPins)
				if newPins != nil && len(newPins) > 0 {
					pinCache[addPath] = append(pinCache[addPath], newPins...)
					//logger.Debugf("Found pins in: %s", addPath)
				}
			}
		} else {
			logger.Debugf("Error with path %s: %s", addPath, err)
		}
	} else {
		logger.Debugf("Error with path %s: %s", addPath, err)
	}

	return pinCache
}

func GetPins(filepath string) []Pin {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	pins := make([]Pin, 0)

	file, err := os.Open(filepath)
	if err == nil {
		pinScanner := bufio.NewScanner(file)

		var crgetRE = regexp.MustCompile(`GET pin\:`)
		var crputRE = regexp.MustCompile(`PUT pin\:`)
		var crEndRE = regexp.MustCompile(`ENDPIN`)
		//var crputEndRE = regexp.MustCompile(`ENDPIN`)

		var foundPinStr string = ""
		var httpVerb string = ""
		for pinScanner.Scan() {
			scanStr := pinScanner.Text()

			matches := crgetRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//pins = append(pins, *NewPin(pinStr))
				if foundPinStr == "" {
					foundPinStr = scanStr
					httpVerb = "GET"
				} else {
					logger.Debugf("Last pin failed, your ENDGET is probably incorrect: %s", foundPinStr)
					break
				}
				//logger.Debugf("GET matches is %d for string: %s", len(matches), foundPinStr)
			}

			matches = crputRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//pins = append(pins, *NewPin(pinStr))
				if foundPinStr == "" {
					foundPinStr = scanStr
					httpVerb = "PUT"
				} else {
					logger.Debugf("Last pin failed, your ENDPUT is probably incorrect: %s", foundPinStr)
					break
				}
				//logger.Debugf("PUT matches is %d for string: %s", len(matches), foundPinStr)
			}

			matches = crEndRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//logger.Debugf("ENDGET matches is %d for string: %s", len(matches), scanStr)
				if foundPinStr != "" {
					newPin := NewPin(httpVerb, foundPinStr)
					newPin = verifyPin(newPin)
					//logger.Debugf("Got pins: %s", pins)
					pins = append(pins, newPin)
					//logger.Debugf("Got pins: %s", pins)
					foundPinStr = ""
				}
			}

			//			matches = crEndRE.FindStringSubmatch(scanStr)
			//			if len(matches) >= 1 {
			//				//logger.Debugf("ENDPUT matches is %d for string: %s", len(matches), scanStr)
			//				if foundPinStr != "" {
			//					newPin := NewPin("PUT", foundPinStr)
			//					newPin = verifyPin(newPin)
			//					//logger.Debugf("Got pins: %s", pins)
			//					pins = append(pins, newPin)
			//					//logger.Debugf("Got pins: %s", pins)
			//					foundPinStr = ""
			//				}
			//			}
		}

		if foundPinStr != "" {
			logger.Debugf("Last pin failed, your last ENDPUT or ENDGET is probably incorrect: %s", foundPinStr)
		}

	} else {
		logger.Debugf("Error with path %s: %s", filepath, err)
	}

	defer file.Close()

	//logger.Debugf("!!!!Got pins: %s", pins)
	return pins
}
