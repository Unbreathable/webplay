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

	// Claim the receiver
	receiver, valid := claimReciever()
	if !valid {
		return c.SendStatus(fiber.StatusConflict)
	}

	return c.JSON(fiber.Map{
		"token": receiver.Token,
	})
}

type Receiver struct {
	Token          string
	currentAttempt *Attempt // The current attempt
}

// The current receiver
var currentReceiver *Receiver = nil

// Returns whether or not the receiver has been claimed
func claimReciever() (*Receiver, bool) {

	// Check if the receiver has already been claimed
	if currentReceiver != nil {
		return nil, false
	}

	// Create a new receiver
	currentReceiver = &Receiver{
		Token: generateToken(12),
	}

	return currentReceiver, true
}
