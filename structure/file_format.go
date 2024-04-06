package structure

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
