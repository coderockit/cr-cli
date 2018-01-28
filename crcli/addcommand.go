package crcli

import (
	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

func AddPaths(args cli.Args) {
	logger := loggo.GetLogger("coderockit.cli.add")
	//logger.Debugf("added file: %s", args.First())

	var pinCache = make(Pinmap)
	ReadInPinCache(pinCache)

	// loop over all of args scanning for cr* directives
	for index, addPath := range args {
		logger.Debugf("Adding path %s at index %d\n", addPath, index)
		pinCache = ProcessPath(addPath, pinCache)
		//logger.Debugf("^^^^^ pinCache is: %s\n", pinCache)
	}

	SavePinCache(pinCache)
}
