package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var ErrAborted = errors.New("aborted")

func BootstrapApp(ctx context.Context, configPath string) (*internal.App, error) {
	wg := ctx.Value(internal.CtxKey("wg")).(*sync.WaitGroup)

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
	wg.Add(1)

	go func() {
		<-ctx.Done()
		store.Close()
		wg.Done()
	}()

	return &internal.App{
		Conf:  conf,
		Codec: codec,
		Store: store,
		Debug: !strings.HasPrefix(internal.Version, "v"),
	}, nil
}
