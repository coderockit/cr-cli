package main

import (
	"os"
	"path/filepath"

	"github.com/coderockit/cr-cli/crcli"
	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

/*
cr add -- adds files that contain CodeRockIt directives that need to be processed
cr diff -- what is changing compared to the CodeRockIt source since the last commit for files that have been added
           this shows diffs for both crget and crput -- you have a unique identifier that never changes
			and a HASH and a human
			readable version number vN.N.N (semver)
cr status -- show list of files that have been added to be processed
cr apply -- process the files that were added and save the changes locally (this step prompts the user
              for input for each CodeRockIt crput directive to collect a comment for the change and a license
			   unless those are listed with the CodeRockIt directive)
** Do not need - cr push -- push the changes up to the CodeRockIt server
** Do not need - cr pull -- pull the changes down from the CodeRockIt server
cr config -- show the configuration contained in the coderockit.json file
*/
func main() {
	var configDir string

	app := cli.NewApp()
	app.Name = "cr"
	app.Usage = "CodeRockIt processor"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Directory, `DIR`, containing the coderockit.json file",
			Destination: &configDir,
		},
	}

	app.Action = func(c *cli.Context) error {
		// loading the config MUST be done first so that logging is configured first
		crcli.LoadConfiguration(configDir)

		logger := loggo.GetLogger("coderockit.cli.main")
		// Create the .coderockit directory in the current directory if it does not exist
		dotcr := filepath.Join(".", ".coderockit")
		if err := os.MkdirAll(dotcr, os.ModePerm); err != nil {
			logger.Debugf("Cannot create the .coderockit directory.")
		}

		// crcli.Hash("hash this")

		return nil
	}

	app.Run(os.Args)

	logger := loggo.GetLogger("coderockit.cli.main")
	logger.Debugf("DONE!!")
}
