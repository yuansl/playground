package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	app := fiber.New(fiber.Config{
		// Prefork:        true,
		BodyLimit:      10 << 20,
		Concurrency:    100,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    10 * time.Second,
		ReadBufferSize: 8 << 10,
	})

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
	println(app.Listen(":3000"))
}
