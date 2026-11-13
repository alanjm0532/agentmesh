package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFiles embed.FS

func registerPages(pages *gin.RouterGroup) {
	files, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	assets := http.FS(files)
	// Serve the directory to avoid net/http redirecting index.html back to /.
	pages.StaticFileFS("/", "/", assets)
	pages.StaticFileFS("/login.html", "login.html", assets)
	pages.StaticFileFS("/favicon.ico", "images/favicon.ico", assets)
	pages.StaticFS("/static", assets)
}
