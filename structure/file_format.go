package structure

import (
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
	rootDir := "/home/amina/Downloads/"
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
