package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
)

type createSenderRequest struct {
	Session webrtc.SessionDescription `json:"session"`
	Token   string                    `json:"token"`
	Code    string                    `json:"code"`
}

// Route: /create_sender
func createSender(c *fiber.Ctx) error {

	// Parse the request
	var req createSenderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Check if the attempt is valid
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if currentReceiver.currentAttempt == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if currentReceiver.currentAttempt.Token != req.Token || currentReceiver.currentAttempt.Challenge != req.Code {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// TODO: Answer the offer

	return c.JSON(fiber.Map{
		"success": true,
	})
}
