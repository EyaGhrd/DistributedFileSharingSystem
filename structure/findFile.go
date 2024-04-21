package structure

import (
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

type Peer struct {
	ID    string
	Files map[byte]byte
}

type Block struct {
	pieces PieceDetails
}

func fileDiscover(host host.Host, dht dht.IpfsDHT) {

}
func (p *Peer) addFile(name byte, content byte) {
	p.Files[name] = content
	fmt.Printf("%s added file %s with content: %s\n", p.ID, name, content)
}
