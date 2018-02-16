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
)

func SavePinsToApply(pinsToApply Pinmap) {
	//FileioLogger.Debugf("Saving pins to apply: %s", pinsToApply)

	pinsToApplyPath := GetPinsToApplyFile()
	jsonString, err := json.Marshal(pinsToApply)
	if err == nil {
		//FileioLogger.Debugf("JSON to save: %s", jsonString)
		err := ioutil.WriteFile(pinsToApplyPath, jsonString, 0644)
		if err != nil {
			FileioLogger.Debugf("Error: %s", err)
		} else {
			FileioLogger.Debugf("Wrote pin cache to: %s", pinsToApplyPath)
		}
	} else {
		FileioLogger.Debugf("Error: %s", err)
	}
}

func ReadInPinsToApply() Pinmap {
	FileioLogger.Debugf("Reading pin cache")

	pinsToApply := make(Pinmap)

	pinsToApplyPath := GetPinsToApplyFile()
	jsonString, err := ioutil.ReadFile(pinsToApplyPath)
	if err == nil {
		err := json.Unmarshal(jsonString, &pinsToApply)
		if err != nil {
			FileioLogger.Debugf("Error unmarshalling json in file %s: %s", pinsToApplyPath, err)
		}
	} else {
		FileioLogger.Debugf("Error reading file %s: %s", pinsToApplyPath, err)
	}

	return pinsToApply
}

func DeleteDirectoryRecursively(pathToDir string) {
	err := os.RemoveAll(pathToDir)

	if err != nil {
		FileioLogger.Debugf("Error deleting directory: %s", err)
	}
}

func DeleteFileOrDir(pathToFile string) {
	err := os.Remove(pathToFile)

	if err != nil {
		FileioLogger.Debugf("Error deleting file or directory: %s", err)
	}
}

func AddPathToPins(addPath string, pinsToApply Pinmap) Pinmap {
	abs, err := filepath.Abs(addPath)
	//FileioLogger.Debugf("Processing path: %s", abs)
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
				//FileioLogger.Debugf("!!!newPins is: %s", newPins)
				if newPins != nil && len(newPins) > 0 {
					pinsToApply[addPath] = append(pinsToApply[addPath], newPins...)
					//FileioLogger.Debugf("Found pins in: %s", addPath)
				}
			}
		} else {
			FileioLogger.Debugf("Error with path %s: %s", addPath, err)
		}
	} else {
		FileioLogger.Debugf("Error with path %s: %s", addPath, err)
	}

	return pinsToApply
}

func RemovePathFromPins(removePath string, pinsToApply Pinmap) Pinmap {
	abs, err := filepath.Abs(removePath)
	if err == nil {
		for pinFile := range pinsToApply {
			if strings.Contains(pinFile, abs) {
				//fmt.Printf("Deleting pins -- key[%s] value[%s]\n", pinFile, pinsToApply[pinFile])
				//pins := pinsToApply[pinFile]
				for _, pin := range pinsToApply[pinFile] {
					versionDir := filepath.Join(GetApplyDirectory(), pin.GroupName, pin.Name, pin.ApplyVersion)
					if pin.ApplyVersion != "" {
						fmt.Printf("Deleting apply cache content %s\n", versionDir)
						DeleteDirectoryRecursively(versionDir)
					}
				}
				delete(pinsToApply, pinFile)
			}
		}
	} else {
		FileioLogger.Debugf("Error with path %s: %s", removePath, err)
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
					FileioLogger.Debugf("Last pin failed, your ENDGET is probably incorrect: %s", foundPinStr)
					break
				}
				//FileioLogger.Debugf("GET matches is %d for string: %s", len(matches), foundPinStr)
			}

			matches = crputRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//pins = append(pins, *NewPin(pinStr))
				if foundPinStr == "" {
					foundPinStr = scanStr
					httpVerb = "PUT"
				} else {
					FileioLogger.Debugf("Last pin failed, your ENDPUT is probably incorrect: %s", foundPinStr)
					break
				}
				//FileioLogger.Debugf("PUT matches is %d for string: %s", len(matches), foundPinStr)
			}

			matches = crputPrivateRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//pins = append(pins, *NewPin(pinStr))
				if foundPinStr == "" {
					foundPinStr = scanStr
					httpVerb = "PUTPRIVATE"
				} else {
					FileioLogger.Debugf("Last pin failed, your ENDPUT is probably incorrect: %s", foundPinStr)
					break
				}
				//FileioLogger.Debugf("PUT matches is %d for string: %s", len(matches), foundPinStr)
			}

			matches = crEndRE.FindStringSubmatch(scanStr)
			if len(matches) >= 1 {
				//FileioLogger.Debugf("ENDGET matches is %d for string: %s", len(matches), scanStr)
				if foundPinStr != "" {
					newPin := NewPin(httpVerb, foundPinStr)
					newPin = verifyPin(newPin, pinContent, false)
					//FileioLogger.Debugf("Got pins: %s", pins)
					pins = append(pins, newPin)
					//FileioLogger.Debugf("Got pins: %s", pins)
					foundPinStr = ""
					pinContent = ""
				}
			}

			if foundPinStr != "" && scanStr != foundPinStr {
				pinContent += scanStr
			}
		}

		if foundPinStr != "" {
			FileioLogger.Debugf("Last pin failed, your last ENDPUT or ENDGET is probably incorrect: %s", foundPinStr)
		}

	} else {
		FileioLogger.Debugf("Error with path %s: %s", pinFile, err)
	}

	defer file.Close()

	//FileioLogger.Debugf("!!!!Got pins: %s", pins)
	return pins
}

func ReadPinContent(rootPath string, pin Pin) (string, string) {
	contentDir := filepath.Join(rootPath, pin.GroupName, pin.Name, pin.ApplyVersion)
	if pin.ApplyVersion != "" {
		pinContentFile := filepath.Join(contentDir, "pinContent.pin")
		pinContent, err := ioutil.ReadFile(pinContentFile)
		if err == nil {
			//fmt.Printf("      >> Content from: %s\n", pinContentFile)
			return pinContentFile, string(pinContent)
		} else {
			FileioLogger.Debugf("Error reading file %s: %s\n", pinContentFile, err)
			return pinContentFile, fmt.Sprintf("%s\n", err)
		}
	}
	return contentDir, fmt.Sprintf("For %s no pin applyVersion, cannot read content!!\n", pin.Verb)
}

func ReadPinContentToGet(pin Pin) (string, string) {
	return ReadPinContent(GetHomeCacheDirectory(), pin)
}

func ReadPinContentToPut(pin Pin) (string, string) {
	return ReadPinContent(GetApplyDirectory(), pin)
}

func WritePinContent(rootPath string, pin Pin, pinContent string) {
	contentDir := filepath.Join(rootPath, pin.GroupName, pin.Name, pin.ApplyVersion)
	if pin.ApplyVersion != "" {
		if err := os.MkdirAll(contentDir, os.ModePerm); err == nil {
			pinContentFile := filepath.Join(contentDir, "pinContent.pin")
			err1 := ioutil.WriteFile(pinContentFile, []byte(pinContent), 0644)
			if err1 == nil {
				pinHashFile := filepath.Join(contentDir, "pinHash.txt")
				err2 := ioutil.WriteFile(pinHashFile, []byte(pin.Hash), 0644)
				if err2 != nil {
					FileioLogger.Debugf("Cannot write to file %s: %s", pinHashFile, err2)
				}
			} else {
				FileioLogger.Debugf("Cannot write to file %s: %s", pinContentFile, err1)
			}
		} else {
			FileioLogger.Debugf("Cannot create the %s directory: %s", contentDir, err)
		}
	} else {
		FileioLogger.Debugf("Fatal: for %s no pin applyVersion, content NOT saved %s:\n%s", pin.Verb, pin, pinContent)
	}
}

func WritePinContentToApply(pin Pin, pinContent string) {
	WritePinContent(GetApplyDirectory(), pin, pinContent)
}

func WritePinApiContentToCache(pin Pin, pinContent string) {
	WritePinContent(GetHomeCacheDirectory(), pin, pinContent)
}

func VerifyGetPinsAgainstLocalPutPins(pinsToApply Pinmap) Pinmap {

	for pinFile := range pinsToApply {
		pins := pinsToApply[pinFile]
		for pinIndex, pin := range pins {
			if pin.IsGet() && !pin.ApiSuccess() {
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
									FileioLogger.Debugf("Found matching versions: %s", matchingVersions)
									pin.ApiMsg = fmt.Sprintf("Local PUT apply cache matches these versions "+
										"%s :: %s -- MAY OVERRIDE FAILURE -- %s", matchingVersions,
										absPinDir, pin.ApiMsg)
									pins[pinIndex] = pin
								} else {
									FileioLogger.Debugf("Cannot verify pin using local apply %s", pin)
								}
							} else {
								FileioLogger.Debugf("Error listing files in directory %s: %s", absPinDir, err)
							}
						}
					} else {
						FileioLogger.Debugf("Cannot verify pin using local apply %s: %s", pin, err)
					}
				} else {
					FileioLogger.Debugf("Cannot verify pin using local apply %s: %s", pin, err)
				}
			}
		}

		pinsToApply[pinFile] = pins
	}

	return pinsToApply
}

func GetEndingPath(fullPath string, startingPath string) string {
	//FileioLogger.Debugf("Starting path: %s", startingPath)
	beginIndex := strings.Index(fullPath, startingPath)
	if beginIndex != -1 {
		return fullPath[beginIndex+len(startingPath)+1:]
	}
	return fullPath
}

func FinishApplyingPut(pin Pin) {
	//	// remove content out of apply cache for PUT
	//	versionDir := filepath.Join(GetApplyDirectory(), pin.GroupName, pin.Name, pin.ApplyVersion)
	//	//fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! 1) Deleting directory %s\n", versionDir)
	//	if pin.ApplyVersion != "" {
	//		//fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!2) Deleting directory %s\n", versionDir)
	//		DeleteDirectoryRecursively(versionDir)
	//	}
}

func FinishApplyingGet(pinFile string, pin Pin) {
	// for GET, put content into applicable file of the project
}
