package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Info         TorrentFileInfo `bencode:"info"`
	Announce     string          `bencode:"announce"`
	AnnounceList [][]string      `bencode:"announce-list"`
	Comment      string          `bencode:"comment"`
	CreationDate uint64          `bencode:"creation date"` // not official element
	CreatedBy    string          `bencode:"created by"`
	Encoding     string          `bencode:"encoding"`
	HttpSeeds    []string        `bencode:"httpseeds"` // not official element
	UrlList      []string        `bencode:"url-list"`  // not official element
}

type TorrentFileInfo struct {
	PieceLength uint32 `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"` //its array of 20 byte sha1 arrays
	Name        string `bencode:"name"`
	FileLength  uint64 `bencode:"length"`
}

func DecodeTorrentFile(torrent_file_reader io.Reader) (*TorrentFile, error) {
	torrent := TorrentFile{}
	err := bencode.Unmarshal(torrent_file_reader, torrent)
	if err != nil {
		return nil, err
	}
	return &torrent, nil
}

// convert info field to sha1 hash
func (i *TorrentFileInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}
func (i *TorrentFileInfo) splitPieceHashes() ([][20]byte, error) {
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
