package p2pnet

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

func BootstrapDHT(ctx context.Context, host host.Host, kad_dht *dht.IpfsDHT) {
	err := kad_dht.Bootstrap(ctx)
	if err != nil {
		fmt.Println("Error while setting the DHT in bootstrap mode:", err)
		return
	}
	fmt.Println("Successfully set the DHT in bootstrap mode")

	for _, bootstrapPeer := range dht.GetDefaultBootstrapPeerAddrInfos() {
		fmt.Println("Attempting to connect to bootstrap peer:", bootstrapPeer.ID, "@", bootstrapPeer.Addrs)
		err := host.Connect(ctx, bootstrapPeer)
		if err != nil {
			fmt.Println("Error while connecting to:", bootstrapPeer.ID)
		} else {
			fmt.Println("Successfully connected to:", bootstrapPeer.ID)
			// Send the successfully connected bootstrap peer to the channel
			//peerChan <- bootstrapPeer
			//fmt.Println("OMG", peerChan)

		}
	}
	fmt.Println("Done with all connections to DHT Bootstrap peers")
}

func InitDHT(ctx context.Context, host host.Host) *dht.IpfsDHT {
	kad_dht, err := dht.New(ctx, host)
	if err != nil {
		fmt.Println("Error while creating the new DHT")
	} else {
		fmt.Println("Successfully created the new DHT")
	}
	return kad_dht
}

//func HandleDHT(ctx context.Context, host host.Host) *dht.IpfsDHT {
//	kad_dht := InitDHT(ctx, host)
//	BootstrapDHT(ctx, host, kad_dht,peerChan)
//	return kad_dht
//}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this to ensure proper CORS handling
	},
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var lock = sync.Mutex{}

func BroadcastPeers(peerChan <-chan PeerUpdate) {
	for peerUpdate := range peerChan {
		peerInfo := fmt.Sprintf("Status: %s, Peer ID: %s, Addresses: %v", peerUpdate.Status, peerUpdate.PeerAddrInfo.ID, peerUpdate.PeerAddrInfo.Addrs)
		broadcastToAllClients(peerInfo)
	}
}

func broadcastToAllClients(message string) {
	lock.Lock()
	defer lock.Unlock()
	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			fmt.Println("Failed to send to a client:", err)
			delete(clients, conn)
		}
	}
}

// wsHandler manages WebSocket connections and sends discovered peers to clients
func wsHandler(peerChan <-chan PeerUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("WebSocket upgrade error:", err)
			return
		}
		fmt.Println("WebSocket connection established")

		go func() {
			for peerUpdate := range peerChan {
				fmt.Println("Sending peer info to client:", peerUpdate.PeerAddrInfo.ID)
				peerData := fmt.Sprintf("Status: %s, Peer ID: %s, Addresses: %v", peerUpdate.Status, peerUpdate.PeerAddrInfo.ID, peerUpdate.PeerAddrInfo.Addrs)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(peerData)); err != nil {
					fmt.Println("Error sending message:", err)
					return
				}
			}
		}()
	}
}

// StartServer initializes the HTTP server and handles incoming requests
func StartServer(peerChan chan PeerUpdate) {
	http.HandleFunc("/ws", wsHandler(peerChan))

	fmt.Println("HTTP server starting on port 8082...")
	err := http.ListenAndServe(":8082", nil) // Capture the error returned by ListenAndServe
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func peersHandler(peerChan chan peer.AddrInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}
func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
