package cli

import (
	"context"
	"fmt"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
	"github.com/salmanmorshed/simplelinkshortener/internal/web"
)

func StartServer(ctx context.Context, cfgPath string) error {
	conf, err := cfg.LoadConfigFromFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := db.NewStore(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer store.Close()

	codec, err := intstrcodec.New(conf.Codec.Alphabet, conf.Codec.BlockSize)
	if err != nil {
		return fmt.Errorf("failed to initialize codec: %w", err)
	}

	serve := web.SetupRouter(ctx, conf, store, codec)

	errCh := make(chan error)
	go func() { errCh <- serve() }()
	go func() {
		<-ctx.Done()
		web.CacheWaitGroup.Wait()
		errCh <- nil
	}()

	return <-errCh
}
