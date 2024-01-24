package server

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed webroot/*
var embedded embed.FS

func ServeEmbeddedWebpage() func(*gin.Context) {
	return func(c *gin.Context) {
		page := strings.ToLower(c.Param("page"))

		file, err := embedded.Open(fmt.Sprintf("webroot/%s.html", page))
		if err != nil {
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				c.String(http.StatusInternalServerError, "Error opening file")
			}
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading file")
			return
		}

		c.Header("Content-Type", "text/html")
		c.Writer.Write(data)
	}
}
