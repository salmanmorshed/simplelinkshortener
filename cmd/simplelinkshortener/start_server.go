package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/server"
)

func startServer(conf *config.Config, db *sqlx.DB) error {
	codec, err := intstrcodec.CreateCodec(conf.Codec.Alphabet, conf.Codec.BlockSize, conf.Codec.MinLength)
	if err != nil {
		return fmt.Errorf("failed to initialize codec: %v", err)
	}

	router := server.CreateRouter(conf, db, codec)
	if conf.Server.UseTLS {
		err = router.RunTLS(
			fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port),
			conf.Server.TLSFiles.Certificate,
			conf.Server.TLSFiles.PrivateKey,
		)
	} else {
		err = router.Run(fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port))
	}
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}
