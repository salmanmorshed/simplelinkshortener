package web

import (
	"context"
	"io/fs"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func SetupRouter(globalCtx context.Context, conf *cfg.Config, store db.Store, codec *intstrcodec.Codec) *gin.Engine {
	var static fs.FS
	if strings.HasPrefix(cfg.Version, "v") {
		static = efs
		gin.SetMode(gin.ReleaseMode)
	} else {
		static = os.DirFS("internal/web")
	}

	handler := Handler{conf, store, codec}

	router := gin.Default()

	if conf.Server.UseCORS {
		router.Use(CORSMiddleware(conf))
	}

	authed := BasicAuthMiddleware(store)

	router.GET("/", handler.OpenHomePage())
	router.GET("/:id", handler.OpenShortLink(globalCtx))
	router.GET("/web", authed, ServeStaticFile(static, "static/index.html"))

	api := router.Group("/api", APIVersionMiddleware(), authed)
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

	return router
}
