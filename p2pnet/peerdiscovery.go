package p2pnet

import (
	"context"
	"fmt"
	"github.com/Mina218/FileSharingNetwork/stream"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

type PeerUpdate struct {
	PeerAddrInfo peer.AddrInfo
	Status       string // "discovered" or "connected"
}

func DiscoverPeers(ctx context.Context, host host.Host, config *Config, kad_dht *dht.IpfsDHT, peerChan chan PeerUpdate) {
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successful in advertising service")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done. Exiting loop.")
			return
		default:
			peerChannel, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
			if err != nil {
				fmt.Println("Error while finding peers for service:", config.RendezvousString, err)
				continue
			}

			for peerAdd := range peerChannel {
				if peerAdd.ID == host.ID() || len(peerAdd.Addrs) == 0 {
					continue
				}

				// Send discovered peer info
				peerChan <- PeerUpdate{PeerAddrInfo: peerAdd, Status: "discovered"}

				fmt.Println("Attempting to connect to:", peerAdd.ID)
				if err := connectToPeer(ctx, host, peerAdd, config.ProtocolID); err != nil {
					fmt.Println("Error connecting to peer:", peerAdd.ID, err)
					continue
				}
				streams,err := host.NewStream(ctx,peerAdd.ID,protocol.ID(config.ProtocolID))
				if err != nil {
					continue
				}
				// Send connected peer info
				fmt.Println("Connected to:", peerAdd.ID)
				peerChan <- PeerUpdate{PeerAddrInfo: peerAdd, Status: "connected"}
				stream.HandleInputStream(streams)
				stream.HandleIncomingStreams(ctx, host)
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func connectToPeer(ctx context.Context, host host.Host, peerAdd peer.AddrInfo, protocolID string) error {
	stream, err := host.NewStream(ctx, peerAdd.ID, protocol.ID(protocolID))
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	return host.Connect(ctx, peerAdd)
}