package main

import (
	"encoding/json"
	"log"
  "sync"
  "os"
	"github.com/pion/webrtc/v3"
	"github.com/gofiber/fiber/v2"
)

var (
	peerConnection *webrtc.PeerConnection
	pcLock         sync.Mutex
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {
	app := fiber.New()


	app.Post("/offer", handleOffer)
	app.Post("/candidate", handleCandidate)
	app.Get("/watch", handleWatch)

	app.Static("/", "./public/")

	log.Fatal(app.Listen(getPort()))
}

func handleOffer(c *fiber.Ctx) error {
	pcLock.Lock()
	defer pcLock.Unlock()

	// Decode the received offer
	var offer webrtc.SessionDescription
	err := json.Unmarshal(c.Body(), &offer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to decode offer"})
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create peer connection"})
	}

	// Set the remote description
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set remote description"})
	}

	// Create an answer and set the local description
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create answer"})
	}

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set local description"})
	}

	// Return the answer to the sharing client
	return c.JSON(answer)
}

func handleCandidate(c *fiber.Ctx) error {
	pcLock.Lock()
	defer pcLock.Unlock()

	// Decode the received ICE candidate
	var iceCandidate webrtc.ICECandidateInit
	err := json.Unmarshal(c.Body(), &iceCandidate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to decode ICE candidate"})
	}

	// Add the ICE candidate to the peer connection
	err = peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add ICE candidate"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func handleWatch(c *fiber.Ctx) error {
	pcLock.Lock()
	defer pcLock.Unlock()

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create offer"})
	}

	return c.JSON(offer)
}



