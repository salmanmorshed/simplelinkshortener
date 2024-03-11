package web

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/salmanmorshed/simplelinkshortener/internal"
	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func CORSMiddleware(conf *cfg.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if conf.Server.UseCORS && origin != "" && slices.Contains(conf.Server.CORSOrigins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Accept-Encoding, Content-Type, Content-Length")
			c.Header("Access-Control-Expose-Headers", "WWW-Authenticate, Content-Type, Content-Length, X-API-Version")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Header("X-API-Version", internal.Version)

		c.Next()
	}
}

func BasicAuthMiddleware(store db.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication credentials"})
			return
		}

		user, err := store.RetrieveUser(c, username)
		if err != nil || !db.ValidatePassword(user.Password, password) {
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
		user := c.MustGet("user").(*db.User)

		if !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not admin"})
			return
		}

		c.Next()
	}
}