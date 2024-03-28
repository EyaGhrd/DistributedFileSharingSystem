package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	_ "github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/libp2p/go-libp2p/core/routing"
	mdns2 "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"

	"log"
	"os"
	"strings"
	"sync"
	"time"
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

func main() {
	ctx := context.Background()

	sourceNode := SourceNode()
	println("-- SOURCE NODE INFORMATION --")
	printNodeID(sourceNode)
	printNodeAddresses(sourceNode)

	targetNode := DestinationNode()
	println("-- TARGET NODE INFORMATION --")
	printNodeID(targetNode)
	printNodeAddresses(targetNode)

	connectToNodeFromSource(sourceNode, targetNode)
	fmt.Printf("##########################\n")

	// view host details and addresses
	fmt.Printf("host ID %s\n", sourceNode.ID())
	fmt.Printf("following are the assigned addresses\n")
	for _, addr := range sourceNode.Addrs() {
		fmt.Printf("%s\n", addr.String())
	}
	fmt.Printf("\n")

	// create a new PubSub service using the GossipSub router

	_, err := pubsub.NewGossipSub(ctx, sourceNode)
	if err != nil {
		panic(err)
	}
	var bootstrapPeers []multiaddr.Multiaddr
	//this are multi-address string
	// the first thing node node whe it join to the network it connect to well-known nodes
	// their name is bootstrap nodes
	Adress := []multiaddr.Multiaddr{
		multiaddrString("/ip4/172.17.0.1/tcp/4001"),
		multiaddrString("/ip4/172.17.0.1/tcp/4000"),
		multiaddrString("/ip4/172.17.0.1/tcp/5000"),
		multiaddrString("/ip4/172.17.0.1/tcp/4500"),
		multiaddrString("/ip4/172.17.0.1/tcp/4600"),
	}
	for _, addr := range Adress {
		node, err := createNodeWithMultiaddr(ctx, addr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("*********\n")
		//so to add them to dht table i should create node with given address
		//and then combine the address with bootstrap node id
		peerAddr := addr.Encapsulate(multiaddrString(fmt.Sprintf("/ipfs/%s", node.ID())))

		// Append the bootstrap peer address to the list
		bootstrapPeers = append(bootstrapPeers, peerAddr)
		// Print the ID and addresses of the created node
		fmt.Println("Node ID:", node.ID())
		fmt.Println("Node Addresses:")
		for _, addr := range node.Addrs() {
			fmt.Println(addr)
		}
		fmt.Println("---------------------------------------")
	}
	// this node a node created with default parameter so when i run code
	// discover function found it
	_ = SourceNode()
	//create dht
	dht, err := NewDh(ctx, sourceNode, bootstrapPeers)
	if err != nil {
		panic(err)
	}
	//the rendezvous is FileSharingNetwork
	go Discoverr(ctx, targetNode, dht, "FileSharingNetwork")

	if err := setupDiscovery(sourceNode); err != nil {
		panic(err)
	}
	ps, err := pubsub.NewGossipSub(context.Background(), sourceNode)
	if err != nil {
		log.Fatal(err)
	}
	room := "FileSharingNetwork"
	topic, err := ps.Join(room)
	if err != nil {
		panic(err)
	}
	publish(ctx, topic)

	fmt.Printf("##########################\n")

	println(fmt.Sprintf("Source node peers: %d", countSourceNodePeers(sourceNode)))
}
func multiaddrString(addr string) multiaddr.Multiaddr {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		panic(err)
	}
	return maddr
}

// this is not necessary now
func publish(ctx context.Context, topic *pubsub.Topic) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Printf("enter message to publish: \n")

			msg := scanner.Text()
			if len(msg) != 0 {
				// publish message to topic
				bytes := []byte(msg)
				topic.Publish(ctx, bytes)
			}
		}
	}
}

type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound now we will try to connect dynamically
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID)
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
}

// DiscoveryServiceTag tag for notifying system
const DiscoveryServiceTag = "FileSharingNetwork-pubsub"
const DiscoveryInterval = time.Hour

func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns2.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

// Discoverr :this function is to discover peer with the help of dht
// dht library it's not only for hashing peer also contain routing
// rendezvous is like the point you find others node (rendezvous point)
func Discoverr(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
	//var disc discovery.Discovery
	config := parseFlags()
	//this is from routing file you can check the routing.go under discovery directory
	routingDiscovery := drouting.NewRoutingDiscovery(dht)
	//same thing
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	fmt.Println("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	fmt.Println("Searching for other peers...")
	//discoveryRouting := routing.NewDiscoveryRouting(disc)

	//_, err2 := routingDiscovery.Advertise(ctx, rendezvous)
	//if err2 != nil {
	//	log.Printf("Error advertising rendezvous: %v", err2)
	//	return
	//}
	//to find peer check routing.go
	peerChan, _ := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	//thee is problem here in the for loop it will give just when peer
	// it need some functionality from time library
	for peer := range peerChan {
		if peer.ID == h.ID() {
			continue
		}

		fmt.Println("peer ID: ", peer.ID, "\nhostID: ", h.ID())
		fmt.Println("Found peer:", peer)

		fmt.Println("Connecting to:", peer)

		fmt.Println("Connected to:", peer)
	}
}

// this function is responsible for configuring the node with command-line
// essentially to specify bunch of characteristic for the node

func parseFlags() *Config {
	c := &Config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "FileSharingNetwork", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/file/1.1.0", "Sets a protocol id for stream headers")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")
	flag.StringVar(&c.dType, "dType", "mdns", "Discovery type")
	flag.Var(&c.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")

	flag.Parse()
	if len(c.BootstrapPeers) == 0 {
		c.BootstrapPeers = dht.DefaultBootstrapPeers
	}

	if err := c.validateConfig(); err != nil {
		panic(err)
	}
	return c
}

type Config struct {
	RendezvousString string
	ProtocolID       string
	BootstrapPeers   addrList
	listenHost       string
	listenPort       int
	dType            string
}
type addrList []multiaddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

// just validation function
func (c *Config) validateConfig() error {
	if c.dType != "mdns" && c.dType != "dht" {
		return fmt.Errorf("Invalid discovery type %v . Please use either 'mdns' or 'dht'", c.dType)
	}
	return nil
}
