package web

import (
	"io/fs"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/salmanmorshed/simplelinkshortener/internal"
)

func SetupRouter(app *internal.App) *gin.Engine {
	var static fs.FS
	if app.Debug {
		static = os.DirFS("internal/web")
	} else {
		static = efs
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(CORSMiddleware(app))

	router.GET("/", OpenHomePage(app))
	router.GET("/:id", OpenShortLink(app))
	router.GET("/web", BasicAuthMiddleware(app), ServeStaticFile(static, "static/index.html"))

	api := router.Group("/api", BasicAuthMiddleware(app))
	api.GET("/links", LinkList(app))
	api.POST("/links", LinkCreate(app))
	api.GET("/links/:id", LinkDetails(app))
	api.DELETE("/links/:id", LinkDelete(app))

	apiAdmin := api.Group("", AdminFilterMiddleware(app))
	apiAdmin.GET("/users", UserList(app))
	apiAdmin.POST("/users", UserCreate(app))
	apiAdmin.GET("/users/:username", UserDetailsOrEdit(app))
	apiAdmin.PATCH("/users/:username", UserDetailsOrEdit(app))
	apiAdmin.DELETE("/users/:username", UserDelete(app))

	return router
}
