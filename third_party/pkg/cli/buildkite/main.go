package buildkite

import (
	"fmt"
	"os"

        "github.com/buildkite/agent/v3/agent"
        "github.com/buildkite/agent/v3/clicommand"
        "github.com/urfave/cli"
)

var AppHelpTemplate = `Usage:

  {{.Name}} <command> [options...]

Available commands are:

  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Use "{{.Name}} <command> --help" for more information about a command.

`

var SubcommandHelpTemplate = `Usage:

  {{.Name}} {{if .VisibleFlags}}<command>{{end}} [options...]

Available commands are:

   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .VisibleFlags}}
Options:

   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

var CommandHelpTemplate = `{{.Description}}

Options:

   {{range .VisibleFlags}}{{.}}
   {{end}}
`

func printVersion(c *cli.Context) {
	fmt.Printf("%v version %v, build %v\n", c.App.Name, c.App.Version, agent.BuildVersion())
}

func Run(args []string) error {
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = "buildkite-agent"
	app.Version = agent.Version()
	app.Commands = []cli.Command{
		clicommand.AgentStartCommand,
		clicommand.AnnotateCommand,
		{
			Name:  "artifact",
			Usage: "Upload/download artifacts from Buildkite jobs",
			Subcommands: []cli.Command{
				clicommand.ArtifactUploadCommand,
				clicommand.ArtifactDownloadCommand,
				clicommand.ArtifactSearchCommand,
				clicommand.ArtifactShasumCommand,
			},
		},
		{
			Name:  "meta-data",
			Usage: "Get/set data from Buildkite jobs",
			Subcommands: []cli.Command{
				clicommand.MetaDataSetCommand,
				clicommand.MetaDataGetCommand,
				clicommand.MetaDataExistsCommand,
				clicommand.MetaDataKeysCommand,
			},
		},
		{
			Name:  "pipeline",
			Usage: "Make changes to the pipeline of the currently running build",
			Subcommands: []cli.Command{
				clicommand.PipelineUploadCommand,
			},
		},
		{
			Name:  "step",
			Usage: "Get or update an attribute of a build step",
			Subcommands: []cli.Command{
				clicommand.StepGetCommand,
				clicommand.StepUpdateCommand,
			},
		},
		clicommand.BootstrapCommand,
	}

	// When no sub command is used
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}

	// When a sub command can't be found
	app.CommandNotFound = func(c *cli.Context, command string) {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}

	if err := app.Run(append([]string{"buildkite-agent"}, args...)); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	return nil
}

