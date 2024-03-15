package web

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var efs embed.FS

func ServeStaticFile(fs fs.FS, relPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := fs.Open(relPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.String(http.StatusNotFound, "Page not found")
			} else {
				c.String(http.StatusInternalServerError, "Failed to open page")
			}
			return
		}

		parts := strings.Split(relPath, ".")
		ext := "." + parts[len(parts)-1]
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			c.String(http.StatusInternalServerError, "Failed to open page")
			return
		}

		c.Header("Content-Type", mimeType)
		if _, err = io.Copy(c.Writer, io.Reader(file)); err != nil {
			c.String(http.StatusInternalServerError, "Failed to load page")
			return
		}
	}
}
