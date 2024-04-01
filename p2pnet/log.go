package p2pnet

import (
	"flag"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"strings"
)

func ParseFlags() *Config {
	c := &Config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "FileSharingNetwork", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.ListenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/file/1.1.0", "Sets a protocol id for stream headers")
	flag.IntVar(&c.ListenPort, "port", 4001, "node listen port")
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
	ListenHost       string
	ListenPort       int
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
