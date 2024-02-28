package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Accept-Encoding, Content-Type, Content-Length")
		c.Header("Access-Control-Expose-Headers", "WWW-Authenticate, Content-Type, Content-Length, X-API-Version")
		c.Header("X-API-Version", config.Version)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func BasicAuthMiddleware(store database.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication credentials"})
			return
		}

		user, err := store.RetrieveUser(username)
		if err != nil || !user.CheckPassword(password) {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "incorrect username and/or password"})
			return
		}

		c.Set("user", user)

		c.Next()
	}
}

func AdminFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*database.User)

		if !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not admin"})
			return
		}

		c.Next()
	}
}
