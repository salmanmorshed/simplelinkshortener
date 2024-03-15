package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var Version = "devel"

type App struct {
	Debug bool
	Conf  *cfg.Config
	Store db.Store
	Codec *intstrcodec.Codec
}

func BootstrapApp(ctx context.Context, configPath string) (*App, error) {
	conf, err := cfg.LoadConfigFromFile(configPath)
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
	go store.Close(ctx)

	return &App{
		Conf:  conf,
		Codec: codec,
		Store: store,
		Debug: !strings.HasPrefix(Version, "v"),
	}, nil
}
