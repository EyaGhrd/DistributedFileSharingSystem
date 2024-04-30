package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"time"
)

var peerChan chan peer.AddrInfo = make(chan peer.AddrInfo, 100) // Adjust buffer size based on expected volume
func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// StartServer starts the HTTP server and returns any errors encountered.
//func StartServer() error {
//	http.HandleFunc("/api/peers", peersHandler)
//	fmt.Println("heuyyyyyyy")
//	return http.ListenAndServe(":8081", nil)
//}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // adjust the origin checking to your requirements
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for peer := range peerChan {
		if err := conn.WriteJSON(peer); err != nil {
			fmt.Println("Error sending peer over WebSocket:", err)
			break
		}
	}
}

func peersHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	select {
	case peer := <-peerChan:
		json.NewEncoder(w).Encode(peer)
	case <-time.After(10 * time.Second): // timeout to prevent hanging indefinitely
		http.Error(w, "No peers available", http.StatusNoContent)
	}
}
