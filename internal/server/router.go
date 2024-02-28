package server

import (
	"embed"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/server/handlers"
)

//go:embed web/*
var efs embed.FS

func CreateRouter(conf *config.Config, store database.Store, codec *intstrcodec.Codec) *gin.Engine {
	if strings.HasPrefix(config.Version, "v") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware())
	}

	router.GET("/", handlers.HomePageHandler(conf))
	router.GET("/:slug", handlers.OpenShortLinkHandler(conf, store, codec))

	router.GET("/web", BasicAuthMiddleware(store), handlers.EmbeddedWebpageHandler(efs, "web/index.html"))

	api := router.Group("/api", BasicAuthMiddleware(store))
	api.GET("/links", handlers.LinkListHandler(store, codec))
	api.POST("/links", handlers.LinkCreateHandler(conf, store, codec))
	api.GET("/links/:slug", handlers.LinkDetailsHandler(store, codec))
	api.DELETE("/links/:slug", handlers.LinkDeleteHandler(store, codec))

	apiAdmin := api.Group("", AdminFilterMiddleware())
	apiAdmin.GET("/users", handlers.UserListHandler(store))
	apiAdmin.POST("/users", handlers.UserCreateHandler(store))
	apiAdmin.GET("/users/:username", handlers.UserDetailsEditHandler(store))
	apiAdmin.PATCH("/users/:username", handlers.UserDetailsEditHandler(store))
	apiAdmin.DELETE("/users/:username", handlers.UserDeleteHandler(store))

	return router
}
