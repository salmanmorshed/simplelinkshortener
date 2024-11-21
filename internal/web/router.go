package web

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func SetupRouter(globalCtx context.Context, conf *cfg.Config, store db.Store, codec *intstrcodec.Codec) func() error {
	handler := Handler{conf, store, codec}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware(conf))
	}

	authed := BasicAuthMiddleware(store)

	router.GET("/", handler.OpenHomePage())
	router.GET("/:id", handler.OpenShortLink(globalCtx))
	router.GET("/web", handler.Frontend())

	router.GET("/api", handler.APIVersion())
	api := router.Group("/api", authed)
	api.GET("/links", handler.LinkList())
	api.POST("/links", handler.LinkCreate())
	api.GET("/links/:id", handler.LinkDetails())
	api.DELETE("/links/:id", handler.LinkDelete())

	apiAdmin := api.Group("", AdminFilterMiddleware())
	apiAdmin.GET("/users", handler.UserList())
	apiAdmin.POST("/users", handler.UserCreate())
	apiAdmin.GET("/users/:username", handler.UserDetailsOrEdit())
	apiAdmin.PATCH("/users/:username", handler.UserDetailsOrEdit())
	apiAdmin.DELETE("/users/:username", handler.UserDelete())

	return func() error {
		if !conf.Server.UseTLS {
			return router.Run(fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port))
		}
		return router.RunTLS(
			fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port),
			conf.Server.TLSCertificate,
			conf.Server.TLSPrivateKey,
		)
	}
}
