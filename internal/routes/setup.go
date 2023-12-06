package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"
	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"gorm.io/gorm"
)

func CreateRouter(conf *config.AppConfig, db *gorm.DB, codec *intstrcodec.CodecConfig) *gin.Engine {
	if !conf.Debug {
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

	admin := private.Group("", AdminFilterMiddleware(db))
	admin.GET("/api/users", UserListHandler(db))
	admin.Any("/api/users/:username", UserManageHandler(db))

	return router
}
