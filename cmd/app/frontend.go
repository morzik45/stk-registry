package main

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/morzik45/stk-registry/frontimport"
)

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func (app *App) initFrontend() {
	// Раздаём vue
	frontend := embedFileSystem{FileSystem: http.FS(frontimport.GetFrontendAssets())}
	staticServer := static.Serve("/", frontend)
	app.router.Use(staticServer)

	// Если путь не найден, перенаправляем на главную
	app.router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet &&
			//!strings.ContainsRune(c.Request.URL.Path, '.') &&
			!strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Request.URL.Path = "/"
			staticServer(c)
		}
	})
}
