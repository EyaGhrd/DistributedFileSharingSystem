package structure

import (
	"context"
	"encoding/hex"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/minio/sha256-simd"
	"io"
	"os"
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
