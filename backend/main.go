package main

//go:generate qtc -dir=templates -ext=html

import (
	"context"
	"crypto/tls"
	"embed"
	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//go:embed dist
var content embed.FS

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	apiGroup := app.Group("/api")

	apiGroup.Post("/save_scheme", func(ctx *fiber.Ctx) error {
		return ctx.SendString("OK")
	})

	apiGroup.Get("/load_scheme", func(ctx *fiber.Ctx) error {
		return ctx.SendString("OK!")
	})
	apiGroup.Get("*", func(ctx *fiber.Ctx) error {
		ctx.Status(404)
		return ctx.SendString("not found route")
	})

	if os.Getenv("ENVIRONMENT") == "production" {
		app.Use(filesystem.New(filesystem.Config{
			Root:       http.FS(content),
			MaxAge:     3600,
			PathPrefix: "/dist",
			Index:      "index.html",
		}))
	} else {
		app.Use(func(c *fiber.Ctx) error {
			return proxy.Do(c, "http://localhost:5173"+c.Path())
		})
	}

	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	wg, _ := errgroup.WithContext(context.Background())
	wg.Go(func() error {
		_ = <-c
		log.Info().Msg("Gracefully shutting down...")
		return app.Shutdown()
	})

	wg.Go(func() error {
		return app.Listen(":8099")
	})

	if err := wg.Wait(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
