package main

import (
	"os"

	"github.com/coderockit/cr-cli/crcli"
	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

/*
cr add -- adds files that contain CodeRockIt directives that need to be processed,
          also scans each file for the crget and crput directives and verifies that
		   that the directives are correct with the CodeRockIt API, if the crget or crput
			directives are incorrect in any way then the file is added as red, if the
			file does not have any crput or crget directives then it is not added
cr remove -- Remove a file or a directory and it's files recursively from the list of files to apply changes to
cr empty -- Empty out all files from the list of files to apply changes to
cr status -- show list of files that have been added to be processed
cr diff -- what is changing compared to the CodeRockIt source since the last commit for files that have been added
           this shows diffs for both crget and crput -- you have a unique identifier that never changes
			and a HASH and a human
			readable version number vN.N.N (semver)
cr apply -m "comment" -- process the files that were added and save the changes locally (this step prompts the user
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

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add or Re-add a file or a directory and it's files recursively",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				crcli.AddPaths(c.Args())
				return nil
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "Remove a file or a directory and it's files recursively from the list of files to apply changes to",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				//logger := loggo.GetLogger("coderockit.cli.main")
				//logger.Debugf("remove files: %s", "need to implement")
				crcli.RemovePaths(c.Args())
				return nil
			},
		},
		{
			Name:    "empty",
			Aliases: []string{"e"},
			Usage:   "Empty out all files from the list of files to apply changes to",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				//logger := loggo.GetLogger("coderockit.cli.main")
				//logger.Debugf("remove all files: %s", "need to implement")
				crcli.EmptyPinsToApply(c.Args())
				return nil
			},
		},
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Show the list of files that an apply will affect",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				//logger := loggo.GetLogger("coderockit.cli.main")
				//logger.Debugf("list of files: %s", "need to implement")
				crcli.ShowStatus(c.Args())
				return nil
			},
		},
		{
			Name:    "diff",
			Aliases: []string{"d"},
			Usage:   "Show the detailed source code diffs of files for changes that would happen on an apply",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				logger := loggo.GetLogger("coderockit.cli.main")
				logger.Debugf("set of diffs: %s", "need to implement")
				return nil
			},
		},
		{
			Name:    "apply",
			Aliases: []string{"y"},
			Usage:   "Apply the changes for files that have been added and then reset the list of files to empty",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				logger := loggo.GetLogger("coderockit.cli.main")
				logger.Debugf("apply changes: %s", "need to implement")
				return nil
			},
		},
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Show the configuration in the coderockit.json file",
			Action: func(c *cli.Context) error {
				crcli.LoadConfiguration(configDir)
				logger := loggo.GetLogger("coderockit.cli.main")
				logger.Debugf("config is: %s", "need to implement")
				return nil
			},
		},
	}

	//	app.Action = func(c *cli.Context) error {
	//		// loading the config MUST be done first so that logging is configured first
	//		crcli.LoadConfiguration(configDir)

	//		logger := loggo.GetLogger("coderockit.cli.main")
	//		logger.Debugf("Still invoking last app.Action function!!!")

	//		// crcli.Hash("hash this")

	//		return nil
	//	}

	app.Run(os.Args)

	logger := loggo.GetLogger("coderockit.cli.main")
	logger.Debugf("DONE!!")
}
