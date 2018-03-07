package main

import (
	"os"

	"github.com/coderockit/cr-cli/crcli"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	var configDir string

	app := cli.NewApp()
	app.Name = "cr"
	app.Usage = "CodeRockIt processor"
	app.Version = "1.1.3"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Directory, `DIR`, containing the coderockit.json file",
			Destination: &configDir,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Start a new project",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.Init(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add or Re-add a file or a directory and it's files recursively",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.AddPaths(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "Remove a file or a directory and it's files recursively from the list of files to apply changes to",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.RemovePaths(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "empty",
			Aliases: []string{"e"},
			Usage:   "Empty out all files from the list of files to apply changes to",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.EmptyPinsToApply(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Show the list of files that an apply will affect",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.ShowStatus(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "diff",
			Aliases: []string{"d"},
			Usage:   "Show the detailed source code diffs for all pins or just a specific pin",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.ShowDiffs(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "apply",
			Aliases: []string{"y"},
			Usage:   "Apply the changes for files that have been added and then remove the pins that were applied successfully",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.ApplyPins(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "show",
			Aliases: []string{"w"},
			Usage:   "Show a list of pins and/or pin versions associated to a given key. The key category is one of: user or group or pin.",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.Show(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "find",
			Aliases: []string{"f"},
			Usage:   "Find a list of pins and/or pin versions whose names and/or content contains the given search string.",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.Find(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Show the configuration in the coderockit.json file",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.ShowConfig(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "perms",
			Aliases: []string{"p"},
			Usage:   "Grant/Remove/Modify permissions for users in groups and pins you manage",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.ApplyPermissions(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "mesg",
			Aliases: []string{"m"},
			Usage:   "Send a message to users who are members of your same groups and pins OR request access to a group or pin",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.SendMessage(c.Args())
				}
				return nil
			},
		},
		{
			Name:    "hash",
			Aliases: []string{"x"},
			Usage:   "Calculate the SHA-512 hash of the content in a given file",
			Action: func(c *cli.Context) error {
				if crcli.LoadConfiguration(c, configDir) {
					crcli.CalculateHash(c.Args())
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
