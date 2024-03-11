package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/salmanmorshed/intstrcodec"
	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var Aborted = errors.New("aborted")

func newAppFromCLI(CLICtx *cli.Context) (*internal.App, error) {
	conf, err := cfg.LoadConfigFromFile(CLICtx.Value("config").(string))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	codec, err := intstrcodec.New(conf.Codec.Alphabet, conf.Codec.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize codec: %w", err)
	}

	store, err := db.NewPgStore(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}
	go store.Close(CLICtx.Context)

	return &internal.App{
		Conf:  conf,
		Codec: codec,
		Store: store,
	}, nil
}

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

	err := CLIApp.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, Aborted) {
			os.Exit(1)
		}
	}
}