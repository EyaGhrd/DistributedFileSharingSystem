package p2pnet

import (
	"bufio"
	"io/ioutil"

	"context"
	"fmt"
	stream "github.com/Mina218/FileSharingNetwork/stream"
	"github.com/Mina218/FileSharingNetwork/structure"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"strings"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

var fileList = make([]string, 0)

type PeerInfo struct {
	Files []string
}

var peerFilesMap = make(map[peer.ID]PeerInfo)

func DiscoverPeers(ctx context.Context, host host.Host, config *Config, kad_dht *dht.IpfsDHT) {
	// Create the routing discovery
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successful in advertising service")
	// Continuously discover and connect to peers
	connectedPeers := make([]peer.ID, 0)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done. Exiting loop.")
			return
		default:
			// Find peers advertising the same rendezvous string
			peerChannel, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
			if err != nil {
				fmt.Println("Error while finding some peers for service:", config.RendezvousString, err)
				continue
			}

			for peerAdd := range peerChannel {
				// Skip if the peer is the same as the current host or has no addresses
				if peerAdd.ID == host.ID() || len(peerAdd.Addrs) == 0 {
					continue
				}

				// Check if the peer's multiaddresses match any of our host's multiaddresses
				isDuplicate := false
				for _, addr := range peerAdd.Addrs {
					for _, ownAddr := range host.Addrs() {
						if addr.Equal(ownAddr) {
							isDuplicate = true
							break
						}
					}
				}

				// If it's a duplicate, skip connecting
				if isDuplicate {
					continue
				}

				// Connect to the peer
				fmt.Println("Found peerAdd:", peerAdd.ID)
				streams, err := host.NewStream(ctx, peerAdd.ID, protocol.ID(config.ProtocolID))
				streamo, err := host.NewStream(ctx, peerAdd.ID, protocol.ID(config.ProtocolID))

				if err != nil {
					//fmt.Println("Connection failed:", err)
					continue
				} else {

				}

				err = host.Connect(ctx, peerAdd)
				if err != nil {
					//fmt.Println("Connecting failed to", peerAdd.ID, ":", err)
					continue
				}
				peerExists := false
				fmt.Println("Connecting to:", peerAdd.ID)
				if err != nil {
					//fmt.Println("Connection failed:", err)
					continue
				} else {
					fmt.Println("Connected to:", peerAdd.ID)
					for _, p := range connectedPeers {
						if p == peerAdd.ID {

							peerExists = true
							break
						}
					}
					if peerExists {
						println("already connected ")
					} else {
						dirPath := "/home/amina/Desktop/FileSharingNetwork/log"
						connectedPeers = append(connectedPeers, peerAdd.ID)
						sendFileList(ctx, host, peerAdd, dirPath, streamo)
						handleFileList(peerAdd.ID, streamo)
						filename := chooseFileAndRequest(ctx, host)
						stream.HandleInputStream(streams, filename)
						stream.HandleIncomingStreams(ctx, host)

					}

				}
			}
		}
		time.Sleep(time.Second * 5) // Adjust the sleep time as needed
	}
}
func sendFileList(ctx context.Context, host host.Host, peerAdd peer.AddrInfo, dirPath string, streams network.Stream) {
	// Retrieve the list of files from the directory
	fileList, err := structure.GetFileListYouHave(dirPath)
	if err != nil {
		fmt.Println("Error getting file list:", err)
		return
	}

	// Prepare the file list message
	fileListMsg := []byte(strings.Join(fileList, "\n"))

	// Open a stream to the peer and send the file list
	_, err = streams.Write(fileListMsg)
	if err != nil {
		fmt.Println("Error sending file list to peer:", err)
		return
	}
}
func RequestIncomingFile(stream network.Stream) {
	defer stream.Close()
	reader := bufio.NewReader(stream)
	_, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	bool := sendORno()
	if bool {
		chooseFile(stream)
	}

}
func handleFileList(peerID peer.ID, stream network.Stream) {
	defer stream.Close()

	// Read the file list from the stream
	reader := bufio.NewReader(stream)
	fileListBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println("Error reading file list:", err)
		return
	}

	// Split the received data into individual file names
	fileList := strings.Split(string(fileListBytes), "\n")

	// Update the peerFilesMap with the received file list

	peerFilesMap[peerID] = PeerInfo{Files: fileList}
}
func chooseFile(stream network.Stream) {

	var filename string
	_, err := fmt.Scanln(&filename)
	if err != nil {
		return
	}
}
func sendORno() bool {
	//to edit logic
	var boolP bool
	fmt.Println("enter yes if you want to send file ender /no if you don't want to send file")
	_, err := fmt.Scanln(boolP)
	if err != nil {
		panic(err)
	}

	return boolP
}
func chooseFileAndRequest(ctx context.Context, host host.Host) string {
	// Assume fileListMap is already populated with peer IDs and their file lists
	// Print available peers and their files
	for peerID, peerInfo := range peerFilesMap {
		fmt.Println("Peer:", peerID)
		fmt.Println("Files:")
		for _, file := range peerInfo.Files {
			fmt.Println(file)
		}
		fmt.Println()
	}

	// Prompt user to choose a peer and a file
	var peerID peer.ID
	fmt.Println("Enter the ID of the peer from which you want to request a file:")
	_, err := fmt.Scanln(&peerID)
	if err != nil {
		fmt.Println("Error reading peer ID:", err)
		println(err)
	}

	var fileName string
	fmt.Println("Enter the name of the file you want to request:")
	_, err = fmt.Scanln(&fileName)
	time.Sleep(time.Second * 50)
	if err != nil {
		fmt.Println("Error reading file name:", err)
		println(err)
	}

	// Check if the chosen peer has the requested file
	peerInfo, ok := peerFilesMap[peerID]
	if !ok {
		fmt.Println("Peer", peerID, "not found or file list not received.")
		println(err)
	}

	var fileFound bool
	for _, file := range peerInfo.Files {
		if file == fileName {
			fileFound = true
			break
		}
	}

	if fileFound {
		// Send a request to the chosen peer for the requested file
		// Implement this part based on your existing sendFileRequest function
		err := sendFileRequest(ctx, host, peerID, fileName)
		if err != nil {
			println(err)
		}
		fmt.Println("Request sent to peer", peerID, "for file", fileName)
	} else {
		fmt.Println("File", fileName, "not found on peer", peerID)
	}
	return fileName
}
func sendFileRequest(ctx context.Context, host host.Host, peerID peer.ID, fileName string) error {
	// Open a stream to the peer
	stream, err := host.NewStream(ctx, peerID, protocol.ID("file-transfer"))
	if err != nil {
		return fmt.Errorf("failed to open stream to peer %s: %s", peerID, err)
	}
	defer stream.Close()

	// Send the file request message
	_, err = fmt.Fprintf(stream, "%s\n", fileName)
	if err != nil {
		return fmt.Errorf("failed to send file request to peer %s: %s", peerID, err)
	}

	return nil
}
