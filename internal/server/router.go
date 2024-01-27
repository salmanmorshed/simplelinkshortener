package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"
	"github.com/salmanmorshed/simplelinkshortener/internal/server/handlers"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
)

func CreateRouter(conf *config.Config, db *sqlx.DB, codec *intstrcodec.Codec) *gin.Engine {
	if config.Version != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware())
	}

	router.GET("/", handlers.HomePageHandler(conf))
	router.GET("/:slug", handlers.OpenShortLinkHandler(conf, db, codec))

	private := router.Group("/private", BasicAuthMiddleware(db))
	private.GET("/:page", ServeEmbeddedWebpage())
	private.GET("/api/links", handlers.LinkListHandler(db, codec))
	private.POST("/api/links", handlers.LinkCreateHandler(conf, db, codec))
	private.GET("/api/links/:slug", handlers.LinkDetailsHandler(db, codec))
	private.DELETE("/api/links/:slug", handlers.LinkDeleteHandler(db, codec))

	admin := private.Group("", AdminFilterMiddleware())
	admin.GET("/api/users", handlers.UserListHandler(db))
	admin.POST("/api/users", handlers.UserCreateHandler(db))
	admin.GET("/api/users/:username", handlers.UserDetailsEditHandler(db))
	admin.PATCH("/api/users/:username", handlers.UserDetailsEditHandler(db))
	admin.DELETE("/api/users/:username", handlers.UserDeleteHandler(db))

	return router
}
