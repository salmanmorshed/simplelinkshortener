package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/salmanmorshed/intstrcodec"

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

	router.GET("/", AppRootHandler(conf))
	router.GET("/:slug", OpenShortUrlHandler(db, codec))

	private := router.Group("/private", BasicAuthMiddleware(db))
	private.GET("/:page", ServeEmbeddedWebpage())
	private.GET("/api/links", LinkListHandler(db, codec))
	private.POST("/api/links", LinkCreateHandler(conf, db, codec))
	private.GET("/api/links/:slug", LinkDetailsHandler(db, codec))
	private.DELETE("/api/links/:slug", LinkDeleteHandler(db, codec))

	admin := private.Group("", AdminFilterMiddleware())
	admin.GET("/api/users", UserListHandler(db))
	admin.POST("/api/users", UserCreateHandler(db))
	admin.GET("/api/users/:username", UserDetailsEditHandler(db))
	admin.PATCH("/api/users/:username", UserDetailsEditHandler(db))
	admin.DELETE("/api/users/:username", UserDeleteHandler(db))

	return router
}
