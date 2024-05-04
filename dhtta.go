package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Mina218/FileSharingNetwork/filetransfer"
	"github.com/Mina218/FileSharingNetwork/p2pnet"
	"github.com/Mina218/FileSharingNetwork/stream"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"log"
	"net/http"
	"path"
	"path/filepath"
)

func createNode(config *p2pnet.Config) (host.Host, error) {
	fmt.Printf("[*] Listening on: %s with port: %d\n", config.ListenHost, config.ListenPort)

	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", config.ListenHost, config.ListenPort))

	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return nil, err
	}

	fmt.Println("host ID: ", host.ID())
	fmt.Println("host address: ", host.Addrs())
	return host, nil
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func hostIDHandler(h host.Host) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"hostID": "%s"}`, h.ID().String())))
	}
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Set the path to the directory where the files are stored
	fileStoragePath := "/home/eya/FileSharingNetwork/log/"

	// Get the filename from the URL parameter
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename not specified", http.StatusBadRequest)
		return
	}

	// Clean the filename to prevent path traversal vulnerabilities
	cleanFilename := filepath.Base(filename)

	// Create the full path to the file
	filePath := path.Join(fileStoragePath, cleanFilename)

	// Check if file exists and open
	file, err := http.Dir(fileStoragePath).Open(cleanFilename)
	if err != nil {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set the header to download the file instead of opening it
	w.Header().Set("Content-Disposition", "attachment; filename="+cleanFilename)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}
func main() {
	peerChan := make(chan p2pnet.PeerUpdate, 100)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	config := p2pnet.ParseFlags()
	h, err := createNode(config)
	if err != nil {
		fmt.Println("Error creating node:", err)
		return
	}
	h.SetStreamHandler(protocol.ID(config.ProtocolID), stream.HandleInputStream)

	kadDht := p2pnet.InitDHT(ctx, h)

	go p2pnet.BootstrapDHT(ctx, h, kadDht)
	go p2pnet.DiscoverPeers(ctx, h, config, kadDht,peerChan)

	go func() {
		http.HandleFunc("/host-id", hostIDHandler(h))
		log.Println("Host ID server starting on port 8090...")
		log.Fatal(http.ListenAndServe(":8090", nil))
	}()

	go func() {
		p2pnet.StartServer(peerChan)
		p2pnet.BroadcastPeers(peerChan)
	}()

	go func() {
		http.HandleFunc("/api/files", filetransfer.EnableCORS(filetransfer.FileHandler))
		log.Println("Server starting on port :8088...")
		log.Fatal(http.ListenAndServe(":8088", nil))
	}()

	go func() {
		http.HandleFunc("/request-file", stream.HandleFileRequest)
		log.Println("Server starting on port :8089...")
		log.Fatal(http.ListenAndServe(":8089", nil))
	}()
	block := make(chan struct{})
	<-block
}
