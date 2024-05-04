package p2pnet

import (
	"context"
	"testing"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
)

func TestDiscoverPeers(t *testing.T) {
	// Create a mock network with 2 nodes
	mn, err := mocknet.FullMeshConnected(2)
	if err != nil {
		t.Fatal(err)
	}

	// Create two fake nodes on the mock network
	node1, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}

	node2, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}

	// Create a Kad-DHT for each node
	kadDHT1, err := dht.New(context.Background(), node1)
	if err != nil {
		t.Fatal(err)
	}

	kadDHT2, err := dht.New(context.Background(), node2)
	if err != nil {
		t.Fatal(err)
	}

	// Create a routing discovery for each Kad-DHT
	routingDiscovery1 := drouting.NewRoutingDiscovery(kadDHT1)
	routingDiscovery2 := drouting.NewRoutingDiscovery(kadDHT2)

	// Advertise the rendezvous service on each routing discovery
	dutil.Advertise(context.Background(), routingDiscovery1, "your-rendezvous-string")
	dutil.Advertise(context.Background(), routingDiscovery2, "your-rendezvous-string")

	// Create a config struct for the DiscoverPeers function
	//config := Config{
	//	RendezvousString: "your-rendezvous-string",
	//	ProtocolID:       "/your/protocol/id",
	//}

	// Call the DiscoverPeers function with the hosts and config as arguments
	//DiscoverPeers(context.Background(), node1, &config, kadDHT1)

	// Check that host1 is connected to host2
	if len(node1.Network().Conns()) != 1 {
		t.Errorf("Expected host1 to be connected to host2, but it is connected to %d peers", len(node1.Network().Conns()))
	}

	// Call the DiscoverPeers function with the hosts and config as arguments
	//DiscoverPeers(context.Background(), node2, &config, kadDHT2)

	// Check that host2 is connected to host1
	if len(node2.Network().Conns()) != 1 {
		t.Errorf("Expected host2 to be connected to host1, but it is connected to %d peers", len(node2.Network().Conns()))
	}
}
