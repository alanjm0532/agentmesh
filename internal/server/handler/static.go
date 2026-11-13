package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFiles embed.FS

func registerStatic(router *gin.Engine) {
	files, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	assets := http.FS(files)
	// Serve the directory to avoid net/http redirecting index.html back to /.
	router.StaticFileFS("/", "/", assets)
	router.StaticFileFS("/login.html", "login.html", assets)
	router.StaticFileFS("/favicon.ico", "images/favicon.ico", assets)
	router.StaticFS("/static", assets)
}
