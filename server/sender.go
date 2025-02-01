package main

import (
	"errors"
	"io"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pion/webrtc/v4"
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

	// Lock the mutex
	currentReceiver.Mutex.Lock()
	defer currentReceiver.Mutex.Unlock()

	// Make sure there is a receiver and no current sender
	if currentReceiver.currentSender != nil {
		currentReceiver.Mutex.Unlock()
		return c.SendStatus(fiber.StatusConflict)
	}

	// Create a new attempt
	attempt := currentReceiver.MakeAttempt(req.Name)

	// Return a new attempt token
	return c.JSON(fiber.Map{
		"token": attempt.Token,
	})
}

type Sender struct {
	Mutex     *sync.Mutex // For not having concurrent writes
	Token     string      // Token to identify the attempt
	Name      string      // Name of the sender
	Challenge string      // Code the sender has to enter
	Accepted  bool        // Whether the attempt has been accepted
	Connected bool        // Whether the sender is connected or not

	// Current sender connection
	connection   *webrtc.PeerConnection
	currentTrack *webrtc.TrackRemote
	localTrack   *webrtc.TrackLocalStaticRTP
}

func (r *Receiver) MakeAttempt(name string) *Sender {
	// Create a new attempt
	attempt := &Sender{
		Mutex:     &sync.Mutex{}, // Initialize mutex
		Token:     generateToken(12),
		Name:      name,
		Challenge: generateNumbers(6),
	}

	// Register the attempt in the receiver
	currentReceiver.currentSender = attempt

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

	currentReceiver.Mutex.Lock()
	defer currentReceiver.Mutex.Unlock()

	// Check if the attempt is valid
	if currentReceiver.currentSender == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}

	sender := currentReceiver.currentSender
	sender.Mutex.Lock()
	defer sender.Mutex.Unlock()

	if sender.Token != req.Token || sender.Challenge != req.Code {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// Change the status of the attempt to accepted
	sender.Accepted = true

	// Send that everything worked, they can now use the code to create a new connection
	return c.SendStatus(fiber.StatusOK)
}

// Route: /sender/connect
func createSenderConnection(c *fiber.Ctx) error {

	// Parse the request
	var req struct {
		Token string                    `json:"token"`
		Offer webrtc.SessionDescription `json:"offer"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Prevent any concurrent modifications
	currentReceiver.Mutex.Lock()
	defer currentReceiver.Mutex.Unlock()

	// Make sure there is a sender
	sender := currentReceiver.currentSender
	if sender == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Prevent any concurrent modifications for sender
	sender.Mutex.Lock()
	defer sender.Mutex.Unlock()

	// Make sure the sender is valid
	if sender.Token != req.Token || !sender.Accepted {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// Create a new connection
	sender = currentReceiver.currentSender
	var err error
	sender.connection, err = api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("error: couldn't create peer connection")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Take in the offer
	if err := sender.connection.SetRemoteDescription(req.Offer); err != nil {
		log.Println("error: couldn't set remote description")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Add the required transceivers
	if _, err := sender.connection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		log.Println("error: couldn't add video transceiver")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Wait for the track to come in
	sender.connection.OnTrack(handleSenderTrack)

	// Handle the connection state
	sender.connection.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		log.Println("sender connection state:", pcs.String())
		sender.Mutex.Lock()
		if pcs == webrtc.PeerConnectionStateConnected {
			sender.Connected = true
		}
		sender.Mutex.Unlock()
	})

	// Create an answer for the client
	answer, err := sender.connection.CreateAnswer(nil)
	if err != nil {
		log.Println("error: couldn't create answer for sender")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := sender.connection.SetLocalDescription(answer); err != nil {
		log.Println("error: couldn't set local description for sender connection")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Wait for the ICE gathering to be completed
	gatherComplete := webrtc.GatheringCompletePromise(sender.connection)
	<-gatherComplete

	return c.JSON(sender.connection.LocalDescription())
}

// Handle the on track event of a sender
func handleSenderTrack(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {

	// Make sure there isn't another track already
	currentReceiver.Mutex.Lock()
	if currentReceiver.currentSender.currentTrack != nil {
		currentReceiver.Mutex.Unlock()
		return
	}
	sender := currentReceiver.currentSender

	// Disconnect the sender when this function returns
	defer func() {
		currentReceiver.currentSender.connection.Close()
		currentReceiver.currentSender = nil
	}()

	// Create a local track for forwarding
	track, err := webrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, tr.ID(), tr.StreamID())
	if err != nil {
		log.Println("Couldn't create track for sender:", err)
		return
	}

	// Set the track
	sender.Mutex.Lock()
	sender.currentTrack = tr
	sender.localTrack = track
	sender.Mutex.Unlock()
	currentReceiver.Mutex.Unlock()

	// Handle all the packets from the sender track
	for {
		// Read RTP packets being sent on the channel
		packet, _, readErr := tr.ReadRTP()
		if readErr != nil {
			log.Println("sender error: couldn't read, closing:", readErr)
			return
		}

		sender.Mutex.Lock()
		if sender.localTrack != nil {
			err := sender.localTrack.WriteRTP(packet)
			sender.Mutex.Unlock()
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				log.Println("sender error: couldn't write to local track, closing:", err)
				return
			}
		} else {
			sender.Mutex.Unlock()
		}
	}
}
