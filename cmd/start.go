package cmd

import (
	"fmt"

	"github.com/salmanmorshed/intstrcodec"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/routes"
	"github.com/urfave/cli/v2"
)

func startServer(c *cli.Context) error {
	var err error

	conf, err := config.LoadConfigFromFile(c.Value("config").(string))
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	db, err := database.CreateGORM(conf)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	codec, err := intstrcodec.CreateCodec(conf.Codec.Alphabet, conf.Codec.BlockSize, conf.Codec.MinLength)
	if err != nil {
		return fmt.Errorf("failed to initialize codec: %v", err)
	}

	router := routes.CreateRouter(conf, db, codec)
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
		return fmt.Errorf("failed to run server: %v", err)
	}

	return nil
}
