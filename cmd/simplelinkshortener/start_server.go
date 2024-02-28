package main

import (
	"fmt"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/server"
)

func startServer(conf *config.Config, store database.Store) error {
	codec, err := intstrcodec.CreateCodec(conf.Codec.Alphabet, conf.Codec.BlockSize, conf.Codec.MinLength)
	if err != nil {
		return fmt.Errorf("failed to initialize codec: %w", err)
	}

	router := server.CreateRouter(conf, store, codec)
	if conf.Server.UseTLS {
		err = router.RunTLS(
			fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port),
			conf.Server.TLSConfig.Certificate,
			conf.Server.TLSConfig.PrivateKey,
		)
	} else {
		err = router.Run(fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port))
	}
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
