package cli

import (
	"context"
	"fmt"

	"github.com/salmanmorshed/simplelinkshortener/internal/web"
)

func StartServer(ctx context.Context, cfgPath string) error {
	app, err := BootstrapApp(ctx, cfgPath)
	if err != nil {
		return err
	}

	router := web.SetupRouter(app)

	if app.Conf.Server.UseTLS {
		err = router.RunTLS(
			fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port),
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
