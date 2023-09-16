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
	conf, err := config.LoadConfigFromFile(c.Value("config").(string))
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	db, err := database.CreateGORM(conf)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	codec, err3 := intstrcodec.CreateCodec(conf.Codec.Alphabet, conf.Codec.BlockSize, conf.Codec.MinLength)
	if err3 != nil {
		return fmt.Errorf("failed to initialize codec: %v", err3)
	}

	router := routes.GetRouter(conf, db, codec)

	var err4 error
	if conf.Server.UseTls {
		err4 = router.RunTLS(
			fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port),
			conf.Server.TlsFiles.Certificate,
			conf.Server.TlsFiles.PrivateKey,
		)
	} else {
		err4 = router.Run(fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port))
	}
	if err4 != nil {
		return fmt.Errorf("failed to run server: %v", err4)
	}
	return nil
}
