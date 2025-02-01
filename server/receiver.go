package main

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
)

// Route: /receiver/create
func createReceiver(c *fiber.Ctx) error {

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
	Mutex         *sync.Mutex // For not having concurrent writes
	Token         string
	currentSender *Sender // The current attempt

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
		Mutex: &sync.Mutex{},
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

	// Make sure there are no concurrent errors
	currentReceiver.Mutex.Lock()
	defer currentReceiver.Mutex.Unlock()

	// Check if there is a current attempt
	if currentReceiver.currentSender != nil {

		// Check if the attempt has been accepted and tell the receiver to wait for a connection in that case
		if currentReceiver.currentSender.Connected {
			return c.JSON(fiber.Map{
				"exists":    true,
				"completed": true,
				"accepted":  currentReceiver.currentSender.Accepted,
				"name":      currentReceiver.currentSender.Name,
				"code":      currentReceiver.currentSender.Challenge,
			})
		}

		return c.JSON(fiber.Map{
			"exists":    true,
			"completed": false,
			"accepted":  currentReceiver.currentSender.Accepted,
			"name":      currentReceiver.currentSender.Name,
			"code":      currentReceiver.currentSender.Challenge,
		})
	}

	// If there isn't one, return that nothing exists
	return c.JSON(fiber.Map{
		"exists": false,
	})
}

// Route: /receiver/connect
func createReceiverConnection(c *fiber.Ctx) error {
	if currentReceiver == nil {
		return c.SendStatus(fiber.StatusBadGateway)
	}

	receiver := currentReceiver
	receiver.Mutex.Lock()
	if receiver.currentSender == nil || receiver.currentSender.localTrack == nil {
		receiver.Mutex.Unlock()
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Store track reference safely
	receiver.currentSender.Mutex.Lock()
	localTrack := receiver.currentSender.localTrack
	receiver.currentSender.Mutex.Unlock()
	receiver.Mutex.Unlock()

	// Parse the request
	var req struct {
		Token string                    `json:"token"`
		Offer webrtc.SessionDescription `json:"offer"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Check the token
	receiver.Mutex.Lock()
	if receiver.Token != req.Token {
		receiver.Mutex.Unlock()
		return c.SendStatus(fiber.StatusBadRequest)
	}
	receiver.Mutex.Unlock()

	// Create a new connection
	var err error
	receiver.connection, err = api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("error: couldn't create peer connection:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Take in the offer
	if err := receiver.connection.SetRemoteDescription(req.Offer); err != nil {
		log.Println("error: couldn't set remote description:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Add the required transceivers
	if _, err := receiver.connection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	}); err != nil {
		log.Println("error: couldn't add video transceiver:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Add the video sender
	rtpSender, err := receiver.connection.AddTrack(localTrack)
	if err != nil {
		log.Println("error: couldn't add local track:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Create an answer for the client
	answer, err := receiver.connection.CreateAnswer(nil)
	if err != nil {
		log.Println("error: couldn't create answer for sender")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := receiver.connection.SetLocalDescription(answer); err != nil {
		log.Println("error: couldn't set local description for sender connection")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Wait for the ICE gathering to be completed
	gatherComplete := webrtc.GatheringCompletePromise(receiver.connection)
	<-gatherComplete

	return c.JSON(*receiver.connection.LocalDescription())
}
