package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
)

// Route: /create_receiver
func createReceiver(c *fiber.Ctx) error {

	// Parse the request
	var req webrtc.SessionDescription
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	return c.SendString("Hello, world!")
}
