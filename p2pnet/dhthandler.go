package p2pnet

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

func BootstrapDHT(ctx context.Context, host host.Host, kad_dht *dht.IpfsDHT, peerChan chan peer.AddrInfo) {
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
			peerChan <- bootstrapPeer
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

func StartServer(peerChan chan peer.AddrInfo) {
	http.HandleFunc("/api/peers", peersHandler(peerChan)) // Traditional HTTP endpoint
	http.HandleFunc("/ws", wsHandler(peerChan))           // WebSocket endpoint

	fmt.Println("HTTP server starting on port 8081...")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func wsHandler(peerChan chan peer.AddrInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		}}
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
