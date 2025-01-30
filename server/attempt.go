package main

import "github.com/gofiber/fiber/v2"

type createAttemptRequest struct {
	Name string `json:"name"`
}

// Route: /create_attempt
func createAttempt(c *fiber.Ctx) error {

	// Parse the request
	var req createAttemptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Make sure there is a receiver
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusConflict)
	}

	// Create a new attempt
	attempt := makeAttempt(req.Name)

	// Return a new attempt token
	return c.JSON(fiber.Map{
		"token": attempt.Token,
	})
}

type Attempt struct {
	Token     string // Token to identify the attempt
	Name      string // Name of the sender
	Challenge string // Code the sender has to enter
	Accepted  bool   // Whether the attempt has been accepted
}

func makeAttempt(name string) *Attempt {

	// Create a new attempt
	attempt := &Attempt{
		Token:     generateToken(12),
		Name:      name,
		Challenge: generateNumbers(6),
	}

	// Register the attempt in the receiver
	currentReceiver.currentAttempt = attempt

	return attempt
}
