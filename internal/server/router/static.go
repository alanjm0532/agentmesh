package router

import (
	"net/http"

	"github.com/alanjm0532/agentmesh/static"
	"github.com/gin-gonic/gin"
)

func registerPages(pages *gin.RouterGroup) {
	assets := http.FS(static.Files)
	// Serve the directory to avoid net/http redirecting index.html back to /.
	pages.StaticFileFS("/", "/", assets)
	pages.StaticFileFS("/login.html", "login.html", assets)
	pages.StaticFileFS("/account.html", "account.html", assets)
	pages.StaticFileFS("/favicon.ico", "images/favicon.ico", assets)
	pages.StaticFS("/static", assets)
}
