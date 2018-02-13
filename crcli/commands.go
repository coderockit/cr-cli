package crcli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

func AddPaths(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.cmds")
	//logger.Debugf("added file: %s", args.First())

	pinsToApply := ReadInPinsToApply()
	logger.Debugf("Found existing pins to apply: %s", pinsToApply)

	// loop over all of args scanning for cr* directives
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
	pinsToApplyPath := GetWorkDirectory() + "/pinsToApply.json"
	DeleteFile(pinsToApplyPath)
}

func ShowStatus(args cli.Args, diffs bool) {
	pinsToApply := ReadInPinsToApply()

	fmt.Println("======================================================" +
		"======================================================" +
		"============================================================")
	for filepath := range pinsToApply {
		// fmt.Printf("key[%s] value[%s]\n", filepath, pinsToApply[filepath])
		//if strings.Contains(filepath, abs) {
		//	delete(pinsToApply, filepath)
		//}
		fmt.Printf("** %s\n", filepath)
		pins := pinsToApply[filepath]
		for _, pin := range pins {
			if strings.HasPrefix(pin.ApiMsg, "Success") {
				fmt.Printf("   -- Ready to apply version %s\n", pin.ApplyVersion)
				fmt.Printf("      ==> %s\n", pin)
			} else {
				fmt.Printf("   -- Cannot apply %s\n", pin)
			}
			fmt.Printf("      >> Api message %s\n", pin.ApiMsg)
		}
		fmt.Println("======================================================" +
			"======================================================" +
			"============================================================")
	}
}
