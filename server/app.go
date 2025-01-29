package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// Configure the basic shit
	app.Use(logger.New())
	app.Use(cors.New())

	// Register all the endpoints
	app.Post("/create_receiver", createReceiver)

	// Start the server
	app.Listen(":3000")
}
