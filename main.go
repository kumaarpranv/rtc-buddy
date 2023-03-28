package main

import (
	"encoding/json"
	"log"
	"net/http"
  "sync"
  "os"
  "time"
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


app.Static("/","./public", fiber.Static{
  Compress:      true,
  ByteRange:     true,
  Browse:        true,
  Index:         "index.html",
  CacheDuration: 10 * time.Second,
  MaxAge:        3600,
})

	app.Listen(getPort())
}

func handleOffer(w http.ResponseWriter, r *http.Request) {
	pcLock.Lock()
	defer pcLock.Unlock()
	// Decode the received offer
	var offer webrtc.SessionDescription
	err := json.NewDecoder(r.Body).Decode(&offer)
	if err != nil {
		http.Error(w, "Failed to decode offer", http.StatusBadRequest)
		return

	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		http.Error(w, "Failed to create peer connection", http.StatusInternalServerError)
		return
	}

	// Set the handler for ICE candidate event
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Send the ICE candidate to the client
			data, err := json.Marshal(candidate.ToJSON())
			if err != nil {
				log.Println("Failed to marshal ICE candidate:", err)
				return
			}
			w.Write(data)
		}
	})

	// Set the remote description
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		http.Error(w, "Failed to set remote description", http.StatusInternalServerError)
		return
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		return
	}

	// Set the local description
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		http.Error(w, "Failed to set local description", http.StatusInternalServerError)
		return
	}

	// Send the answer to the client
	data, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, "Failed to marshal answer", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleCandidate(w http.ResponseWriter, r *http.Request) {
	pcLock.Lock()
	defer pcLock.Unlock()
	// Decode the received ICE candidate
	var candidate webrtc.ICECandidateInit
	err := json.NewDecoder(r.Body).Decode(&candidate)
	if err != nil {
		http.Error(w, "Failed to decode ICE candidate", http.StatusBadRequest)
		return
	}

	// Add the ICE candidate to the peer connection
	err = peerConnection.AddICECandidate(candidate)
	if err != nil {
		http.Error(w, "Failed to add ICE candidate", http.StatusInternalServerError)
		return
	}
}
