package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nrednav/cuid2"
	"uk.co.dudmesh.propolis/internal/boot"
	"uk.co.dudmesh.propolis/internal/handlers"
	"uk.co.dudmesh.propolis/internal/service/user"
)

type UserService interface {
	handlers.UserService
}

type config struct {
	boot.Config
	userService UserService
}

func (c *config) UserService() UserService {
	return c.userService
}

func newConfig(bootConfig *boot.Config) *config {
	userService, err := user.New(bootConfig)
	if err != nil {
		log.Fatalf("creating user service: %+v", err)
	}

	return &config{*bootConfig, userService}
}

func main() {
	bootConfig, err := boot.Load()
	if err != nil {
		log.Fatalf("boot: %+v", err)
	}

	config := newConfig(bootConfig)

	server := echo.New()
	server.Use(middleware.BodyLimit("100M"))
	server.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return cuid2.Generate()
		},
	}))
	server.Use(echoprometheus.NewMiddleware("propolis"))
	server.Use(middleware.Recover())

	server.Logger.SetLevel(log.INFO)

	headers := []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     headers,
		AllowCredentials: true,
	}))

	server.POST("/ingest", handlers.Ingest(config.userService))
	server.GET("/user/:userAddress/publickey", handlers.CreateUser(config.userService))
	server.POST("/local/user", handlers.CreateUser(config.userService))

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
