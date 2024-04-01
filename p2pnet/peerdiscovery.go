package p2pnet

import (
	"context"
	"fmt"

	plog "github.com/Mina218/FileSharingNetwork/fileshare"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func DiscoverPeers(ctx context.Context, host host.Host, config *Config, kad_dht *dht.IpfsDHT) {

	constat := plog.OpenConnectionStatusLog()
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successful in advertising service")
	var connectedPeers []peer.AddrInfo

	//isAlreadyConnected := false
	for len(connectedPeers) < 100 {
		fmt.Fprintln(constat, "Currently connected to", len(connectedPeers), "out of 100 [for service", config.RendezvousString, "]")
		fmt.Fprintln(constat, "TOTAL CONNECTIONS : ", len(host.Network().Conns()))
		peerChannel, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
		if err != nil {
			fmt.Println("Error while finding some peers for service :", config.RendezvousString, err)
		} else {
			fmt.Fprintln(constat, "Successful in finding some peers")
		}

		for peer := range peerChannel {
			if peer.ID == host.ID() {
				continue
			}

			fmt.Println("peer ID: ", peer.ID, "\nhostID: ", host.ID())
			fmt.Println("Found peer:", peer)

			fmt.Println("Connecting to:", peer)
			//_, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

			// Add the discovered peer to connectedPeers
			connectedPeers = append(connectedPeers, peer)

			// Handle the connection logic here

			if err != nil {
				fmt.Println("Connection failed:", err)
				continue
			} else {

			}

			fmt.Println("Connected to:", peer)
		}
	}

}
