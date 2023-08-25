package main

import (
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nrednav/cuid2"
	"uk.co.dudmesh.propolis/internal/boot"
)

type Template struct {
	templates *template.Template
	watcher   *fsnotify.Watcher
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *Template) Watch() {
	var err error

	t.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("watcher: %+v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-t.watcher.Events:
				if !ok {
					return
				}
				log.Infof("event: %+v", event)
				if event.Has(fsnotify.Write) {
					log.Infof("modified file: %s", event.Name)
					t.templates = template.Must(template.ParseGlob("ui/views/*.html"))
				}
			case err, ok := <-t.watcher.Errors:
				if !ok {
					return
				}
				log.Errorf("watcher: %+v", err)
			}
		}
	}()

	// Add a path.
	err = t.watcher.Add("./ui/views")
	if err != nil {
		log.Fatalf("watcher: %+v", err)
	}
}

func (t *Template) Close() {
	if t.watcher != nil {
		t.watcher.Close()
	}
}

func NewTemplate() (*Template, error) {
	t := &Template{
		templates: template.Must(template.ParseGlob("ui/views/*.html")),
	}
	return t, nil
}

func main() {
	config, err := boot.Load()
	if err != nil {
		log.Fatalf("boot: %+v", err)
	}

	server := echo.New()
	server.Use(middleware.BodyLimit("100M"))
	server.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return cuid2.Generate()
		},
	}))
	server.Use(echoprometheus.NewMiddleware("SecureTX"))
	server.Use(middleware.Recover())

	server.Logger.SetLevel(log.INFO)

	// headers := []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID}
	// if !appState.IsProduction() {
	// 	headers = append(headers, SkipRecaptchaHeader)
	// }

	// server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins:     appState.ServerOrigins(),
	// 	AllowHeaders:     headers,
	// 	AllowCredentials: true,
	// }))

	server.Static("/static", "ui/static")

	t, _ := NewTemplate()
	defer t.Close()
	if config.IsDevelopment() {
		t.Watch()
	}
	server.Renderer = t

	server.GET("/", func(c echo.Context) error {
		//return c.String(http.StatusOK, "Hello, World!")
		return c.Render(http.StatusOK, "app.html", nil)
	})

	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := server.Start(":8080"); err != nil && err != http.ErrServerClosed {
			server.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		server.Logger.Fatal(err)
	}
}
