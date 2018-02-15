package crcli

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

func AddPaths(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.cmds")
	//logger.Debugf("added file: %s", args.First())

	pinsToApply := ReadInPinsToApply()
	logger.Debugf("Found existing pins to apply: %s", pinsToApply)

	// loop over all of args scanning for pin directives
	for index, addPath := range args {
		abs, err := filepath.Abs(addPath)
		if err == nil {
			logger.Debugf("Adding path %s at index %d\n", abs, index)
			pinsToApply = AddPathToPins(abs, pinsToApply)
			//logger.Debugf("^^^^^ pinsToApply is: %s\n", pinsToApply)
		} else {
			logger.Debugf("Could not add path %s due to error: %s", addPath, err)
		}
	}

	pinsToApply = VerifyGetPinsAgainstLocalPutPins(pinsToApply)
	SavePinsToApply(pinsToApply)
}

func ApplyPins(args cli.Args) {
	//logger := loggo.GetLogger("coderockit.cli.cmds")
	//logger.Debugf("added file: %s", args.First())

	pinsToApply := ReadInPinsToApply()
	//logger.Debugf("Found existing pins to apply: %s", pinsToApply)

	onlyDoPuts := true
	for {
		for pinFile := range pinsToApply {
			//logger.Debugf("Applying pins in file: %s\n", pinFile)
			pins := pinsToApply[pinFile]
			var failedPins []Pin
			for _, pin := range pins {
				if (onlyDoPuts && pin.IsPut()) || (!onlyDoPuts && pin.IsGet()) {

					pinContent := ""
					//pinContentFile := ""
					if pin.IsPut() {
						_, pinContent = ReadPinContentToPut(pin)
					} else if pin.IsGet() {
						_, pinContent = ReadPinContentToGet(pin)
					}
					pin = verifyPin(pin, pinContent, true)

					//logger.Debugf("pin.ApiMsg: %s", pin.ApiMsg)
					if pin.ApiSuccess() {
						if pin.IsPut() {
							FinishApplyingPut(pin)
						} else if pin.IsGet() {
							FinishApplyingGet(pinFile, pin)
						}
					} else {
						failedPins = append(failedPins, pin)
					}
				} else {
					failedPins = append(failedPins, pin)
				}
			}

			if len(failedPins) > 0 {
				pinsToApply[pinFile] = failedPins
			} else {
				delete(pinsToApply, pinFile)
			}
		}

		if !onlyDoPuts {
			break
		} else {
			onlyDoPuts = false
		}
	}

	SavePinsToApply(pinsToApply)
}

func RemovePaths(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.cmds")

	pinsToApply := ReadInPinsToApply()
	logger.Debugf("Found existing pins to apply: %s", pinsToApply)

	for index, removePath := range args {
		abs, err := filepath.Abs(removePath)
		if err == nil {
			logger.Debugf("Removing path %s at index %d\n", abs, index)
			pinsToApply = RemovePathFromPins(abs, pinsToApply)
			//logger.Debugf("^^^^^ pinsToApply is: %s\n", pinsToApply)
		} else {
			logger.Debugf("Could not remove path %s due to error: %s", removePath, err)
		}
	}

	SavePinsToApply(pinsToApply)

}

func EmptyPinsToApply(args cli.Args) {
	DeleteFileOrDir(GetPinsToApplyFile())
	DeleteDirectoryRecursively(GetApplyDirectory())
}

func ShowStatus(args cli.Args, diffs bool) {
	//logger := loggo.GetLogger("coderockit.cli.cmds")
	pinsToApply := ReadInPinsToApply()

	fmt.Println("======================================================" +
		"======================================================" +
		"============================================================")
	for pinFile := range pinsToApply {
		// fmt.Printf("key[%s] value[%s]\n", pinFile, pinsToApply[pinFile])
		//if strings.Contains(pinFile, abs) {
		//	delete(pinsToApply, pinFile)
		//}
		fmt.Printf("** %s\n", pinFile)
		pins := pinsToApply[pinFile]
		for _, pin := range pins {
			if pin.ApiSuccess() {
				fmt.Printf("   -- Ready to apply version: '%s'\n", pin.ApplyVersion)
				fmt.Printf("      ==> %s\n", pin)
			} else {
				fmt.Printf("   -- Cannot apply version: '%s'\n", pin.ApplyVersion)
				fmt.Printf("      ==> %s\n", pin)
			}
			fmt.Printf("      >> Api message ==> %s\n", pin.ApiMsg)
			//logger.Debugf("Showing diffs: %s", strconv.FormatBool(diffs))
			if diffs {
				if pin.IsPut() {
					pinContentFile, pinContent := ReadPinContentToPut(pin)
					fmt.Printf("      >> Content from: %s\n", pinContentFile)
					fmt.Print(pinContent)
				} else if pin.IsGet() {
					pinContentFile, pinContent := ReadPinContentToGet(pin)
					fmt.Printf("      >> Content from: %s\n", pinContentFile)
					fmt.Print(pinContent)
				}
			}
		}
		fmt.Println("======================================================" +
			"======================================================" +
			"============================================================")
	}
}

func ShowConfig(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.cmds")
	// read in the config file and write it out to the console
	filename, _ := GetConfigFilename()
	fmt.Printf("Using config file: %s\n", filename)
	config, err := ioutil.ReadFile(filename)
	if err == nil {
		fmt.Printf("%s", config)
	} else {
		logger.Debugf("Error reading file %s: %s", filename, err)
	}

	filename = GetHomeConfigFile()
	fmt.Printf("Using home config file: %s\n", filename)
	config, err = ioutil.ReadFile(filename)
	if err == nil {
		fmt.Printf("%s", config)
	} else {
		logger.Debugf("Error reading file %s: %s", filename, err)
	}
}

func CalculateHash(args cli.Args) {
	fileContents, err := ioutil.ReadFile(args[0])
	if err == nil {
		hash := Hash(string(fileContents))
		fmt.Printf("%s\n", hash)
	}
}
