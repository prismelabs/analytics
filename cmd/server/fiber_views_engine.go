package main

import (
	"io/fs"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/embedded"
)

func ProvideFiberViewsEngine(cfg config.Config) fiber.Views {
	viewsFs, err := fs.Sub(embedded.Views, "views")
	if err != nil {
		panic(err)
	}
	engine := html.NewFileSystem(http.FS(viewsFs), ".html")

	if cfg.Server.Debug {
		engine = html.New("internal/embedded/views", ".html")
		engine.Reload(true)
		engine.Debug(true)
	}

	engine.AddFuncMap(sprig.FuncMap())

	return engine
}
