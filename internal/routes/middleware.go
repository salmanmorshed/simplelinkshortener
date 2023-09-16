package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"gorm.io/gorm"
)

func BasicAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, err := database.AuthenticateUser(db, username, password)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
