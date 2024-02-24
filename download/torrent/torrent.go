package torrent

import (
	"io"
	"math/rand"
	"os"
	"time"
)

const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PeerStatus int

const (
	PeerStatusConnecting   PeerStatus = iota // Peer is in the process of establishing a connection
	PeerStatusHandshake                      // Handshake initiated, waiting for peer's handshake
	PeerStatusBitfield                       // Handshake completed, waiting for peer's bitfield
	PeerStatusChoked                         // Peer has choked the connection, no data exchange
	PeerStatusInterested                     // Peer is interested in our data, waiting to unchoke
	PeerStatusUnchoked                       // Peer has been unchoked, data exchange allowed
	PeerStatusRequesting                     // Requesting pieces from peer
	PeerStatusDownloading                    // Downloading data from peer
	PeerStatusSeeding                        // Uploading data to peer
	PeerStatusDisconnected                   // Connection has been terminated
)

type TorrentStatus int

const (
	TorrentStatusPaused TorrentStatus = iota
	TorrentStatusCompleted
	TorrentStatusDownloading
)

type PieceProgress struct {
	Index               uint32
	Buffer              []byte
	DownloadedByteCount uint32
	RequestedByteCount  uint32
	Length              uint32
	Hash                [20]byte // pub requested: u32,
}

type Peer struct {
	Address     PeerAddress
	FullAddress string
	Status      PeerStatus
	Country     string
}

type PeerAddress struct {
	Ip   string
	Port uint16
}

type Torrent struct {
	OurPeerId            string
	TorrentFile          TorrentFile
	InfoHash             string   // hash of info field.
	InfoHashHex          [20]byte //decoded hexadecimal representation of info field hash in bytes
	Bitfield             []byte
	PieceProgresses      []PieceProgress
	Status               TorrentStatus
	DownloadedPieceCount uint32
	TotalPieceCount      uint32
	Peers                map[string]Peer
}

func CreateNewTorrent(torrentFilePath string) (*Torrent, error) {
	rand.Seed(time.Now().UnixNano())

	// Open the torrent raw_file
	raw_file, err := os.Open(torrentFilePath)
	if err != nil {
		return nil, err
	}
	defer raw_file.Close()

	var reader io.Reader = raw_file
	torrentFile, err := DecodeTorrentFile(reader)

	if err != nil {
		return nil, err
	}

	peerIdHead := "-DN0001-"
	// Create a random number generator
	peerIdRngString := make([]byte, 12)
	for _, i := range peerIdRngString {
		peerIdRngString[i] = alphanumericCharset[rand.Intn(len(alphanumericCharset))]
	}
	//append with head
	ourPeerId := peerIdHead + string(peerIdRngString)

	bitfield := make([]byte, 0, len(torrentFile.Info.Pieces)/8)

	//convert info field to sha1 hash
	infoHash, err := torrentFile.Info.hash()
	if err != nil {
		return nil, err
	}

	//convert info field to sha1 hash
	pieceHashes, err := torrentFile.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}

	pieceProgresses := make([]PieceProgress, len(pieceHashes))
	for i, hash := range pieceHashes {
		pieceProgresses = append(pieceProgresses, PieceProgress{
			Buffer:              make([]byte, 0, torrentFile.Info.PieceLength),
			Index:               uint32(i),
			DownloadedByteCount: 0,
			RequestedByteCount:  0,
			Length:              torrentFile.Info.PieceLength,
			Hash:                hash,
		})
	}

	// Create a buffered reader from the torrent file
	torrent := Torrent{
		OurPeerId:   ourPeerId,
		Bitfield:    bitfield,
		InfoHashHex: infoHash,
	}
	return &torrent, nil
}
