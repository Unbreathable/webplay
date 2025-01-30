package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pion/webrtc/v4"
)

var api *webrtc.API

func main() {
	app := fiber.New()

	// Configure the basic shit
	app.Use(logger.New())
	app.Use(cors.New())

	// Register all the endpoints
	app.Post("/receiver/create", createReceiver)
	app.Post("/receiver/check_state", checkReceiverState)
	app.Post("/sender/create", createSender)
	app.Post("/sender/attempt", checkAttempt)

	// Create a new setting engine
	/*
		engine := webrtc.SettingEngine{}

		// Set the port
		mux, err := ice.NewMultiUDPMuxFromPort(5000)
		if err != nil {
			log.Fatal("Couldn't create port multiplexer for the SFU:", err)
		}
		engine.SetICEUDPMux(mux)

		// Create the api using the settings engine
		api = webrtc.NewAPI(webrtc.WithSettingEngine(engine))
	*/

	// Start the server
	app.Listen("127.0.0.1:3000")
}
