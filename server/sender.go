package main

import (
	"github.com/gofiber/fiber/v2"
)

// Route: /sender/create
func createSender(c *fiber.Ctx) error {

	// Parse the request
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Make sure there is a receiver
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Make sure there is no attempt already
	if currentReceiver.curretnSender != nil {
		return c.SendStatus(fiber.StatusConflict)
	}

	// Create a new attempt
	attempt := makeAttempt(req.Name)

	// Return a new attempt token
	return c.JSON(fiber.Map{
		"token": attempt.Token,
	})
}

type Sender struct {
	Token     string // Token to identify the attempt
	Name      string // Name of the sender
	Challenge string // Code the sender has to enter
	Accepted  bool   // Whether the attempt has been accepted
}

func makeAttempt(name string) *Sender {

	// Create a new attempt
	attempt := &Sender{
		Token:     generateToken(12),
		Name:      name,
		Challenge: generateNumbers(6),
	}

	// Register the attempt in the receiver
	currentReceiver.curretnSender = attempt

	return attempt
}

// Route: /sender/attempt
func checkAttempt(c *fiber.Ctx) error {

	// Parse the request
	var req struct {
		Token string `json:"token"`
		Code  string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Check if the attempt is valid
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if currentReceiver.curretnSender == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if currentReceiver.curretnSender.Token != req.Token || currentReceiver.curretnSender.Challenge != req.Code {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// Send that everything worked, they can now use the code to create a new connection
	return c.SendStatus(fiber.StatusOK)
}
