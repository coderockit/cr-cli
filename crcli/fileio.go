package crcli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/juju/loggo"
)

func SavePinsToApply(pinsToApply Pinmap) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	//logger.Debugf("Saving pins to apply: %s", pinsToApply)

	pinsToApplyPath := GetWorkDirectory() + "/pinsToApply.json"
	jsonString, err := json.Marshal(pinsToApply)
	if err == nil {
		//logger.Debugf("JSON to save: %s", jsonString)
		err := ioutil.WriteFile(pinsToApplyPath, jsonString, 0644)
		if err != nil {
			logger.Debugf("Error: %s", err)
		} else {
			logger.Debugf("Wrote pin cache to: %s", pinsToApplyPath)
		}
	} else {
		logger.Debugf("Error: %s", err)
	}
}

func ReadInPinsToApply(pinsToApply Pinmap) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	logger.Debugf("Reading pin cache")
}

func ProcessPath(addPath string, pinsToApply Pinmap) Pinmap {
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
						pinsToApply = ProcessPath(abs+"/"+nextFile.Name(), pinsToApply)
					}
				}
			case mode.IsRegular():
				// do file stuff
				//fmt.Println("file")
				newPins := GetPins(addPath)
				//logger.Debugf("!!!newPins is: %s", newPins)
				if newPins != nil && len(newPins) > 0 {
					pinsToApply[addPath] = append(pinsToApply[addPath], newPins...)
					//logger.Debugf("Found pins in: %s", addPath)
				}
			}
		} else {
			logger.Debugf("Error with path %s: %s", addPath, err)
		}
	} else {
		logger.Debugf("Error with path %s: %s", addPath, err)
	}

	return pinsToApply
}

// dropCR drops a terminal \r from the data.
//func dropCR(data []byte) []byte {
//	if len(data) > 0 && data[len(data)-1] == '\r' {
//		return data[0 : len(data)-1]
//	}
//	return data
//}

// ScanLines is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		//return i + 1, dropCR(data[0:i]), nil
		return i + 2, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		//return len(data), dropCR(data), nil
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func GetPins(filepath string) []Pin {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	pins := make([]Pin, 0)

	file, err := os.Open(filepath)
	if err == nil {
		pinScanner := bufio.NewScanner(file)

		// NOTE: it may be necessary to pass in a binary friendly Split function
		pinScanner.Split(ScanLines)

		var crgetRE = regexp.MustCompile(`GET \/pin\/`)
		var crputRE = regexp.MustCompile(`PUT \/pin\/`)
		var crputPrivateRE = regexp.MustCompile(`PUTPRIVATE \/pin\/`)
		var crEndRE = regexp.MustCompile(`ENDPIN`)
		//var crputEndRE = regexp.MustCompile(`ENDPIN`)

		var foundPinStr string = ""
		var httpVerb string = ""
		var pinContent string = ""

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

			matches = crputPrivateRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//pins = append(pins, *NewPin(pinStr))
				if foundPinStr == "" {
					foundPinStr = scanStr
					httpVerb = "PUTPRIVATE"
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
					newPin = verifyPin(newPin, pinContent)
					//logger.Debugf("Got pins: %s", pins)
					pins = append(pins, newPin)
					//logger.Debugf("Got pins: %s", pins)
					foundPinStr = ""
					pinContent = ""
				}
			}

			if foundPinStr != "" {
				pinContent += scanStr
			}
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
