package decoding

import (
	"bytes"
	"crypto/sha1"
	"downite/download/torrent/peer"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Info         TorrentFileInfo `bencode:"info"`
	Announce     string          `bencode:"announce"`
	AnnounceList [][]string      `bencode:"announce-list"`
	Comment      string          `bencode:"comment"`
	CreationDate int             `bencode:"creation date"` // not official element
	CreatedBy    string          `bencode:"created by"`
	Encoding     string          `bencode:"encoding"`
	HttpSeeds    []string        `bencode:"httpseeds"` // not official element
	UrlList      []string        `bencode:"url-list"`  // not official element
}

type TorrentFileInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"` //its array of 20 byte sha1 arrays
	Name        string `bencode:"name"`
	FileLength  int    `bencode:"length"`
}

func DecodeTorrentFile(torrent_file_reader io.Reader) (*TorrentFile, error) {
	torrent := &TorrentFile{}
	err := bencode.Unmarshal(torrent_file_reader, torrent)
	if err != nil {
		return nil, err
	}
	return torrent, nil
}

// convert info field to sha1 hash
func (i *TorrentFileInfo) Hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}
func (i *TorrentFileInfo) SplitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

// Unmarshal parses peer IP addresses and ports from a buffer
func UnmarshalPeers(peersBin []byte) ([]peer.PeerAddress, error) {
	const peerSize = 6 // 4 for IP, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := fmt.Errorf("Received malformed peers")
		return nil, err
	}
	peers := make([]peer.PeerAddress, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].Ip = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	}
	return peers, nil
}
