package main

import (
	"log"

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

	// WebRTC stuff
	connection *webrtc.PeerConnection
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

// Endpoint for the receiver to check for a code
// Route: /receiver/check_state
func checkReceiverState(c *fiber.Ctx) error {

	// Parse the request
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Check if the token provided is correct
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusBadGateway)
	}
	if currentReceiver.Token != req.Token {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// Check if there is a current attempt
	if currentReceiver.currentAttempt != nil {

		// Check if the attempt has been accepted and tell the receiver to wait for a connection in that case
		if currentReceiver.currentAttempt.Accepted {
			return c.JSON(fiber.Map{
				"exists":    true,
				"completed": true,
				"name":      currentReceiver.currentAttempt.Name,
				"code":      currentReceiver.currentAttempt.Challenge,
			})
		}

		return c.JSON(fiber.Map{
			"exists":    true,
			"completed": false,
			"name":      currentReceiver.currentAttempt.Name,
			"code":      currentReceiver.currentAttempt.Challenge,
		})
	}

	// If there isn't one, return that nothing exists
	return c.JSON(fiber.Map{
		"exists": false,
	})
}

// Endpoint for the receiver to actually create the webrtc connection (incomplete, needs lots of work)
func createReceiverConnection(c *fiber.Ctx) error {

	// Get the current receiver
	receiver := currentReceiver
	if receiver == nil {
		return c.SendStatus(fiber.StatusBadGateway)
	}

	// Parse the request
	var req webrtc.SessionDescription
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Create a new connection
	var err error
	receiver.connection, err = api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("error: couldn't create peer connection")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Take in the offer
	if err := receiver.connection.SetRemoteDescription(req); err != nil {
		log.Println("error: couldn't set remote description")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Add the required transceivers
	if _, err := receiver.connection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		log.Println("error: couldn't add video transceiver")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

// Create the video track for the receiver
func (r *Receiver) createVideoTrack() {

}
