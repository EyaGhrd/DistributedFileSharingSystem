package p2pnet

import (
	"context"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func DiscoverPeers(ctx context.Context, host host.Host, config *Config, kad_dht *dht.IpfsDHT) {
	// Create the routing discovery
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successful in advertising service")

	// Continuously discover and connect to peers
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
				//stream, err := host.NewStream(ctx, peerAdd.ID, protocol.ID(config.ProtocolID))
				if err != nil {
					fmt.Println("Connection failed:", err)
					continue
				} else {

					//go writeDataToStream(stream)
					//go readDataFromStream(stream)
				}

				err = host.Connect(ctx, peerAdd)
				if err != nil {
					fmt.Println("Connecting failed to", peerAdd.ID, ":", err)
					continue
				}

				fmt.Println("Connecting to:", peerAdd.ID)

				// Handle the connection logic here
				if err != nil {
					fmt.Println("Connection failed:", err)
					continue
				} else {
					fmt.Println("Connected to:", peerAdd.ID)
				}
			}
		}
		time.Sleep(time.Second * 5) // Adjust the sleep time as needed
	}
}
