package structure

import (
	"context"
	"encoding/hex"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/minio/sha256-simd"
	"io"
	"os"
	"path/filepath"
)

type TorrentFile struct {
	Name   string         `json:"name"`
	Size   int64          `json:"size"`
	Pieces []PieceDetails `json:"pieces"`
}

// PieceDetails represents details about a single piece of the file.
type PieceDetails struct {
	Index int    `json:"index"`
	Hash  string `json:"hash"`
}

func ListFiles() ([]string, error) {
	rootDir := "/home/amina/Downloads/code"
	var listFiles []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			listFiles = append(listFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return listFiles, nil

}

//	func indexFile(dht *dht.IpfsDHT, ctx context.Context, host host.Host, file TorrentFile) error {
//		//jsonBytes, err := json.Marshal(file)
//		//if err != nil {
//		//	fmt.Println("Error:", err)
//		//	return err
//		//}
//		//if err := dht.Provide(ctx, []byte(file.Name), dht.WithValue(jsonBytes)); err != nil {
//		//	return fmt.Errorf("error providing file metadata in DHT: %v", err)
//		//}
//		hash, err := calculateFileHash(file)
//		if err != nil {
//			panic(err)
//		}
//		return nil
//	}
func IndexFiles(dht *dht.IpfsDHT, ctx context.Context, file TorrentFile) error {
	filePath := file.Name
	hash, err := calculateFileHash(filePath)
	if err != nil {
		return err
	}

	valueBytes := []byte(hash)
	if err := dht.PutValue(ctx, file.Name, valueBytes); err != nil {
		return fmt.Errorf("error putting value in DHT: %v", err)
	}
	return nil
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum), nil
}
