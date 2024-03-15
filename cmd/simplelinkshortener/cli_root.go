package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/urfave/cli/v2"
)

var Aborted = errors.New("aborted")

func main() {
	CLIApp := &cli.App{
		Usage:     "Create a personal link shortening service",
		ArgsUsage: " ",
		Version:   internal.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "config.yaml",
				Usage: "Path to config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "init",
				Usage:    "Initialize a config file",
				Category: "Configuration",
				Action:   initConfigHandler,
			},
			{
				Name:     "start",
				Usage:    "Start the web server",
				Category: "Server",
				Action:   startServerHandler,
			},
			{
				Name:     "useradd",
				Usage:    "Add a new user",
				Category: "User management",
				Action:   addUserHandler,
			},
			{
				Name:     "usermod",
				Usage:    "Modify username or password",
				Category: "User management",
				Action:   modifyUserHandler,
			},
			{
				Name:     "userdel",
				Usage:    "Delete a user",
				Category: "User management",
				Action:   deleteUserHandler,
			},
		},
		HideHelpCommand: true,
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)
	go func() {
		<-sigintCh
		fmt.Println("cleaning up...")
		cancel()

		// wait for database and cache backend to cleanly close
		// todo: implement proper signaling instead of sleep
		time.Sleep(time.Second)

		os.Exit(0)
	}()

	err := CLIApp.RunContext(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, Aborted) {
			os.Exit(1)
		}
	}
}

func getApp(CLICtx *cli.Context) (*internal.App, error) {
	return internal.BootstrapApp(
		CLICtx.Context,
		CLICtx.Value("config").(string),
	)
}
