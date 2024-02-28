package server

import (
	"embed"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/server/handlers"
)

//go:embed web/*
var efs embed.FS

func CreateRouter(conf *config.Config, db *sqlx.DB, codec *intstrcodec.Codec) *gin.Engine {
	if strings.HasPrefix(config.Version, "v") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware())
	}

	router.GET("/", handlers.HomePageHandler(conf))
	router.GET("/:slug", handlers.OpenShortLinkHandler(conf, db, codec))

	router.GET("/web", BasicAuthMiddleware(db), handlers.EmbeddedWebpageHandler(efs, "web/index.html"))

	api := router.Group("/api", BasicAuthMiddleware(db))
	api.GET("/links", handlers.LinkListHandler(db, codec))
	api.POST("/links", handlers.LinkCreateHandler(conf, db, codec))
	api.GET("/links/:slug", handlers.LinkDetailsHandler(db, codec))
	api.DELETE("/links/:slug", handlers.LinkDeleteHandler(db, codec))

	apiAdmin := api.Group("", AdminFilterMiddleware())
	apiAdmin.GET("/users", handlers.UserListHandler(db))
	apiAdmin.POST("/users", handlers.UserCreateHandler(db))
	apiAdmin.GET("/users/:username", handlers.UserDetailsEditHandler(db))
	apiAdmin.PATCH("/users/:username", handlers.UserDetailsEditHandler(db))
	apiAdmin.DELETE("/users/:username", handlers.UserDeleteHandler(db))

	return router
}
