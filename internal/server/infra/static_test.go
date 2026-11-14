package infra

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alanjm0532/agentmesh/static"
	"github.com/gin-gonic/gin"
)

func TestSinglePage(t *testing.T) {
	router := gin.New()
	registerPages(router.Group("/"))
	document, err := static.Files.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"/", "/?page=login", "/?page=account", "/?page=cli-login",
		"/?page=cli-login&user_code=550e8400-e29b-41d4-a716-446655440000",
		"/?page=login&next=%2F%3Fpage%3Dcli-login%26user_code%3D550e8400-e29b-41d4-a716-446655440000",
		"/static/",
	} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("unexpected content type: %q", response.Header().Get("Content-Type"))
			}
			if !bytes.Equal(response.Body.Bytes(), document) {
				t.Fatal("page did not serve the shared HTML document")
			}
		})
	}
}

func TestSeparateStaticAssets(t *testing.T) {
	router := gin.New()
	registerPages(router.Group("/"))
	for path, file := range map[string]string{
		"/favicon.ico":                               "images/favicon.ico",
		"/static/images/logo.png":                    "images/logo.png",
		"/static/images/icons.svg":                   "images/icons.svg",
		"/static/css/app.css":                        "css/app.css",
		"/static/vendor/bootstrap/bootstrap.min.css": "vendor/bootstrap/bootstrap.min.css",
		"/static/js/main.js":                         "js/main.js",
		"/static/js/components/auth-panel.js":        "js/components/auth-panel.js",
	} {
		t.Run(path, func(t *testing.T) {
			asset, err := static.Files.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			if !bytes.Equal(response.Body.Bytes(), asset) {
				t.Fatal("did not serve the separate asset")
			}
		})
	}
}
