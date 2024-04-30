package structure

import (
	"context"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"os"
	"path/filepath"
)

const (
	ListProtocol = "/sukun/file/list"
	GetProtocol  = "/sukun/file/fetch"
	peerChanSize = 100
)

type Peer struct {
	ID    string
	Files map[byte]byte
}

type Block struct {
	pieces PieceDetails
}

//	func ListFiles(rootDir string, dht *dht.IpfsDHT, peer peer.ID) ([]string, error) {
//		var listFiles []string
//		err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
//			if err != nil {
//				return err
//			}
//			if !info.IsDir() {
//				listFiles = append(listFiles, path)
//				filename := info.Name()
//				err := hashfile(filename, dht, peer)
//				if err != nil {
//					return err
//				}
//
//			}
//			return nil
//		})
//		if err != nil {
//			return nil, err
//		}
//		return listFiles, nil
//	}
//
// // hashfile function
//
//	func hashfile(filename string, dht *dht.IpfsDHT, p peer.ID) error {
//		ctx := context.Background()
//
//		// Read file content
//
//		//pkkey := routing.KeyForPublicKey(p)
//		filenameBytes := []byte(filename)
//
//		// Put the file content into the DHT with filename as the key
//		err := dht.PutValue(ctx, DHT_NAMESPACE+filename, filenameBytes)
//		if err != nil {
//			return err
//		}
//		value, err := dht.GetValue(ctx, filename)
//		println("value", value)
//		if err != nil {
//			return err
//		}
//
//		fmt.Println("Successfully added file to DHT")
//
//		return err
//	}
//
// const DHT_NAMESPACE = "/ipns/"
//
// // FindFile function
//
//	func FindFile(filename string, dht *dht.IpfsDHT) (string, error) {
//		ctx := context.Background()
//
//		// Get the value from the DHT using the filename as the key
//		fileContent, err := dht.GetValue(ctx, filename)
//		if err != nil {
//			return "", err
//		}
//
//		// Check if the retrieved content matches the expected file content
//		// You can print or log the content for verification
//		fmt.Println("Retrieved file content:", string(fileContent))
//
//		return string(fileContent), nil
//	}
type Node struct {
	host   host.Host
	dht    *dht.IpfsDHT
	ctx    context.Context
	cancel func()

	dirPath string
}

func GetFileListYouHave(dir string) ([]string, error) {
	var fileList []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}
