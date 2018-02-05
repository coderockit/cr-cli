package crcli

import (
	"path/filepath"

	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

func AddPaths(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.add")
	//logger.Debugf("added file: %s", args.First())

	var pinsToApply = make(Pinmap)
	ReadInPinsToApply(pinsToApply)

	// loop over all of args scanning for cr* directives
	for index, addPath := range args {
		abs, err := filepath.Abs(addPath)
		if err == nil {
			logger.Debugf("Adding path %s at index %d\n", abs, index)
			pinsToApply = ProcessPath(abs, pinsToApply)
			//logger.Debugf("^^^^^ pinsToApply is: %s\n", pinsToApply)
		} else {
			logger.Debugf("Could not add path %s due to error: %s", addPath, err)
		}
	}

	SavePinsToApply(pinsToApply)
}
