package crcli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/juju/loggo"
)

func SavePinsToApply(pinsToApply Pinmap) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	//logger.Debugf("Saving pins to apply: %s", pinsToApply)

	pinsToApplyPath := filepath.Join(GetWorkDirectory(), "pinsToApply.json")
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

func ReadInPinsToApply() Pinmap {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	logger.Debugf("Reading pin cache")

	pinsToApply := make(Pinmap)

	pinsToApplyPath := filepath.Join(GetWorkDirectory(), "pinsToApply.json")
	jsonString, err := ioutil.ReadFile(pinsToApplyPath)
	if err == nil {
		err := json.Unmarshal(jsonString, &pinsToApply)
		if err != nil {
			logger.Debugf("Error unmarshalling json in file %s: %s", pinsToApplyPath, err)
		}
	} else {
		logger.Debugf("Error reading file %s: %s", pinsToApplyPath, err)
	}

	return pinsToApply
}

func DeleteFile(pathToFile string) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	err := os.Remove(pathToFile)

	if err != nil {
		logger.Debugf("Error deleting file: %s", err)
	}
}

func AddPathToPins(addPath string, pinsToApply Pinmap) Pinmap {
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
						pinsToApply = AddPathToPins(filepath.Join(abs, nextFile.Name()), pinsToApply)
					}
				}
			case mode.IsRegular():
				// do file stuff
				//fmt.Println("file")
				newPins := GetPins(addPath)
				delete(pinsToApply, addPath)
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

func RemovePathFromPins(removePath string, pinsToApply Pinmap) Pinmap {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	abs, err := filepath.Abs(removePath)
	if err == nil {
		for pinFile := range pinsToApply {
			// fmt.Printf("key[%s] value[%s]\n", pinFile, pinsToApply[pinFile])
			if strings.Contains(pinFile, abs) {
				delete(pinsToApply, pinFile)
			}
		}
	} else {
		logger.Debugf("Error with path %s: %s", removePath, err)
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
		return i + 1, data[:i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		//return len(data), dropCR(data), nil
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func GetPins(pinFile string) []Pin {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	pins := make([]Pin, 0)

	file, err := os.Open(pinFile)
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

			if foundPinStr != "" && scanStr != foundPinStr {
				pinContent += scanStr
			}
		}

		if foundPinStr != "" {
			logger.Debugf("Last pin failed, your last ENDPUT or ENDGET is probably incorrect: %s", foundPinStr)
		}

	} else {
		logger.Debugf("Error with path %s: %s", pinFile, err)
	}

	defer file.Close()

	//logger.Debugf("!!!!Got pins: %s", pins)
	return pins
}

func ReadPinContentToGet(pin Pin) string {
	return "content\n"
}

func ReadPinContentToPut(pin Pin) string {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	contentDir := filepath.Join(GetApplyDirectory(), pin.GroupName, pin.Name, pin.ApplyVersion)
	if pin.ApplyVersion != "" {
		pinContentFile := filepath.Join(contentDir, "pinContent.pin")
		pinContent, err := ioutil.ReadFile(pinContentFile)
		if err == nil {
			fmt.Printf("      >> Content from: %s\n", pinContentFile)
			return string(pinContent)
		} else {
			logger.Debugf("Error reading file %s: %s", pinContentFile, err)
			return fmt.Sprintf("%s", err)
		}
	}
	return "For PUT no pin applyVersion, cannot read content!!"
}

func WritePinContentToApply(pin Pin, pinContent string) {
	logger := loggo.GetLogger("coderockit.cli.fileio")
	contentDir := filepath.Join(GetApplyDirectory(), pin.GroupName, pin.Name, pin.ApplyVersion)
	if pin.ApplyVersion != "" {
		if err := os.MkdirAll(contentDir, os.ModePerm); err == nil {
			pinContentFile := filepath.Join(contentDir, "pinContent.pin")
			err1 := ioutil.WriteFile(pinContentFile, []byte(pinContent), 0644)
			if err1 == nil {
				pinHashFile := filepath.Join(contentDir, "pinHash.txt")
				err2 := ioutil.WriteFile(pinHashFile, []byte(pin.Hash), 0644)
				if err2 != nil {
					logger.Debugf("Cannot write to file %s: %s", pinHashFile, err2)
				}
			} else {
				logger.Debugf("Cannot write to file %s: %s", pinContentFile, err1)
			}
		} else {
			logger.Debugf("Cannot create the %s directory: %s", contentDir, err)
		}
	} else {
		logger.Debugf("Fatal: for PUT no pin applyVersion, losing content %s:\n%s", pin, pinContent)
	}
}

func VerifyGetPinsAgainstLocalPutPins(pinsToApply Pinmap) Pinmap {
	logger := loggo.GetLogger("coderockit.cli.crcli")

	for pinFile := range pinsToApply {
		pins := pinsToApply[pinFile]
		for pinIndex, pin := range pins {
			if pin.IsGet() {
				// now look in the local ./.coderockit/apply folder to see if
				// a local unapplied pin will match the groupName/pinName/version
				// and if it does then update the pin.ApiMsg to inform the user
				dotcrApply := GetApplyDirectory()
				pinDir := filepath.Join(dotcrApply, pin.GroupName, pin.Name)
				absPinDir, err := filepath.Abs(pinDir)
				if err == nil {
					fi, err := os.Stat(absPinDir)
					if err == nil {
						if fi.IsDir() {
							allFiles, err := ioutil.ReadDir(absPinDir)
							if err == nil {

								var putApplyVersions []string
								for _, nextFile := range allFiles {
									if nextFile.IsDir() {
										//filepath.Join(absPinDir, nextFile.Name())
										putApplyVersions = append(putApplyVersions, nextFile.Name())
									}
								}

								matchingVersions := GetMatchingVersions(pin.Version, putApplyVersions)
								if len(matchingVersions) > 0 {
									logger.Debugf("Found matching versions: %s", matchingVersions)
									pin.ApiMsg = fmt.Sprintf("Local PUT apply cache matches these versions "+
										"%s :: %s -- MAY OVERRIDE FAILURE -- %s", matchingVersions,
										absPinDir, pin.ApiMsg)
									pins[pinIndex] = pin
								} else {
									logger.Debugf("Cannot verify pin using local apply %s", pin)
								}
							} else {
								logger.Debugf("Error listing files in directory %s: %s", absPinDir, err)
							}
						}
					} else {
						logger.Debugf("Cannot verify pin using local apply %s: %s", pin, err)
					}
				} else {
					logger.Debugf("Cannot verify pin using local apply %s: %s", pin, err)
				}
			}
		}

		pinsToApply[pinFile] = pins
	}

	return pinsToApply
}
