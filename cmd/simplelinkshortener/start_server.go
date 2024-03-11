package main

import (
	"fmt"
	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal/web"
)

func startServerHandler(CLICtx *cli.Context) error {
	app, err := newAppFromCLI(CLICtx)
	if err != nil {
		return err
	}

	router := web.CreateRouter(app.Conf, app.Store, app.Codec)

	if app.Conf.Server.UseTLS {
		err = router.RunTLS(
			fmt.Sprintf("%s:%s", app.Conf.Server.Host, app.Conf.Server.Port),
			app.Conf.Server.TLSCertificate,
			app.Conf.Server.TLSPrivateKey,
		)
	} else {
		err = router.Run(fmt.Sprintf("%s:%s", app.Conf.Server.Host, app.Conf.Server.Port))
	}
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}