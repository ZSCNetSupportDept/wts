package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"zsxyww.com/wts/config"
)

var version = "dev"

func main() {
	app := newApp()
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(app.ErrWriter, err)
		os.Exit(1)
	}
}

func newApp() *cli.Command {
	app := &cli.Command{
		Name:                  "wtstool",
		Usage:                 "微信报修系统的管理工具",
		Version:               version,
		EnableShellCompletion: true,
		Writer:                os.Stdout,
		ErrWriter:             os.Stderr,
		ExitErrHandler:        func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "配置文件的路径（YAML格式）",
				Sources: cli.EnvVars("WTS_CONFIG"),
			},
		},
		Commands: []*cli.Command{
			newWXCommand(),
			newJWTCommand(),
		},
		Action: groupAction,
	}
	_ = app.Walk(func(command *cli.Command) error {
		command.OnUsageError = func(ctx context.Context, cmd *cli.Command, err error, isSubcommand bool) error {
			return err
		}
		return nil
	})
	return app
}

func groupAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Present() {
		return fmt.Errorf("unknown command %q", cmd.Args().First())
	}
	if cmd == cmd.Root() {
		return cli.ShowRootCommandHelp(cmd)
	}
	return cli.ShowSubcommandHelp(cmd)
}

func loadToolConfig(cmd *cli.Command) (*config.Config, error) {
	path := cmd.String("config")
	if path == "" {
		return nil, fmt.Errorf("provide a configuration file with --config or WTS_CONFIG")
	}
	return config.LoadFile(path)
}

func rejectExtraArgs(cmd *cli.Command) error {
	if cmd.Args().Present() {
		return fmt.Errorf("unexpected argument %q", cmd.Args().First())
	}
	return nil
}
