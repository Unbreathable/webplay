package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
)

// Route: /create_sender
func createSender(c *fiber.Ctx) error {

	// Parse the request
	var req webrtc.SessionDescription
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}
