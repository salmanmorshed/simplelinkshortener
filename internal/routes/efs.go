package routes

import (
	"io"
	"net/http"

	"embed"

	"github.com/gin-gonic/gin"
)

//go:embed web/*
var embeddedFileSystem embed.FS

func ServeEmbeddedFile(file string, contentType string) func(*gin.Context) {
	return func(c *gin.Context) {
		file, err := embeddedFileSystem.Open(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening file")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading file")
			return
		}

		c.Header("Content-Type", contentType)

		c.Writer.Write(data)
	}
}
