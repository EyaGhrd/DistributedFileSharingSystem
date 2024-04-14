package main

import (
	"context"
	"crypto/rand"
	"github.com/libp2p/go-libp2p/core/crypto"
	//"crypto/rand"
	"fmt"
	"github.com/Mina218/FileSharingNetwork/p2pnet"
	"github.com/libp2p/go-libp2p"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"strings"
	"sync"
)

func SourceNode() host.Host {
	node, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	return node
}
func NewDh(ctx context.Context, host host.Host, Peers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	var options []dht.Option

	if len(Peers) == 0 {
		options = append(options, dht.Mode(dht.ModeServer))
	}

	thisdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}
	if err = thisdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range Peers {
		peerinformations, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			return nil, err
		}
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinformations); err != nil {
				log.Printf("Error while connecting to node %q: %-v", peerinformations, err)
			} else {
				log.Printf("Connection established with bootstrap node: %q", *peerinformations)
			}
		}()
	}
	wg.Wait()

	return thisdht, nil
}
func DestinationNode() host.Host {

	listenAddr := "/ip4/172.17.0.1/tcp/9090"
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddr))
	if err != nil {
		panic(err)
	}

	return node
}
func connectToNodeFromSource(sourceNode host.Host, targetNode host.Host) {
	targetNodeAddressInfo := host.InfoFromHost(targetNode)
	err := sourceNode.Connect(context.Background(), *targetNodeAddressInfo)
	if err != nil {
		panic(err)
	}
}
func countSourceNodePeers(sourceNode host.Host) int {
	return len(sourceNode.Network().Peers())
}
func printNodeID(host host.Host) {
	println(fmt.Sprintf("ID: %s", host.ID().String()))
}

func printNodeAddresses(host host.Host) {
	addressesString := make([]string, 0)
	for _, address := range host.Addrs() {
		addressesString = append(addressesString, address.String())
	}

	println(fmt.Sprintf("Multiaddresses: %s", strings.Join(addressesString, ", ")))
}
func createNodeWithMultiaddr(ctx context.Context, listenAddress multiaddr.Multiaddr) (host.Host, error) {
	// Create a new libp2p node specifying the listen address
	node, err := libp2p.New(libp2p.ListenAddrStrings(listenAddress.String()))
	if err != nil {
		return nil, err
	}
	return node, nil
}

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
func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//r := rand.Reader
	//	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	config := p2pnet.ParseFlags()

	h, err := createNode(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Host created. ID:", h.ID())
	//h.SetStreamHandler(protocol.ID(config.ProtocolID), p2pnet.HandleStream)

	//// Set up a DHT for peer discovery
	kad_dht := p2pnet.InitDHT(ctx, h)
	p2pnet.BootstrapDHT(ctx, h, kad_dht)

	p2pnet.DiscoverPeers(ctx, h, config, kad_dht)

	// Index each file in the directory

	// Wait for shutdown signal do not shutdown by your self danger

	//filePaths, err := structure.ListFiles()
	//if err != nil {
	//	fmt.Println("Error scanning file system:", err)
	//	return
	//}
	//
	//fmt.Println("Files found:")
	//for _, path := range filePaths {
	//	fmt.Println(path)
	//}
	//host.SetStreamHandler(protocol.ID(config.ProtocolID), p2pnet.HandleStream)

}
func multiaddrString(addr string) multiaddr.Multiaddr {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		panic(err)
	}
	return maddr
}

//	type Discovery interface {
//		initDiscovery(host host.Host, config *p2pnet.Config) error
//	}
//type discoveryNotifee struct {
//	PeerChan chan peer.AddrInfo
//}

// interface to be called when new  peer is found
