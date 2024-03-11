package web

import (
	"embed"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
	"github.com/salmanmorshed/simplelinkshortener/internal/web/handlers"
)

//go:embed static/*
var efs embed.FS

func CreateRouter(conf *cfg.Config, store db.Store, codec *intstrcodec.Codec) *gin.Engine {
	if strings.HasPrefix(internal.Version, "v") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(CORSMiddleware(conf))

	router.GET("/", handlers.HomePageHandler(conf))
	router.GET("/:slug", handlers.OpenShortLinkHandler(store, codec))

	router.GET("/web", BasicAuthMiddleware(store), handlers.EmbeddedWebpageHandler(efs, "static/index.html"))

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