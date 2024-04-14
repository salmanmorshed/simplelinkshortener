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

	codec, err := intstrcodec.New(conf.Codec.Alphabet, conf.Codec.BlockSize)
	if err != nil {
		return fmt.Errorf("failed to initialize codec: %w", err)
	}

	router := web.SetupRouter(ctx, conf, store, codec)

	errCh := make(chan error)
	go func() {
		if conf.Server.UseTLS {
			errCh <- router.RunTLS(
				fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port),
				conf.Server.TLSCertificate,
				conf.Server.TLSPrivateKey,
			)
		} else {
			errCh <- router.Run(fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port))
		}
	}()
	go func() {
		<-ctx.Done()
		web.CacheWaitGroup.Wait()
		store.Close()
		errCh <- nil
	}()

	return <-errCh
}
