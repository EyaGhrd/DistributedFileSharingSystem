
package stream


import (
	"fmt"


	"github.com/libp2p/go-libp2p/core/network"
)


func HandleInputStream(stream network.Stream) {
	fmt.Println("New outcome stream detected")
	go SendToStream(stream)
}
