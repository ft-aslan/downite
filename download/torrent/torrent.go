package torrent

import (
	"downite/download/torrent/peer"
	"downite/download/torrent/tracker"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"
)

const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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

type Torrent struct {
	OurPeerId            [20]byte
	TorrentFile          TorrentFile
	InfoHash             [20]byte // hash of info field.
	InfoHashHex          [20]byte //decoded hexadecimal representation of info field hash in bytes
	Bitfield             []byte
	PieceProgresses      []PieceProgress
	Status               TorrentStatus
	DownloadedPieceCount uint32
	TotalPieceCount      uint32
	Length               uint32
	PieceLength          uint32
	Peers                map[string]peer.Peer
}

func New(torrentFilePath string) (*Torrent, error) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Open the torrent rawFile
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
		peerIdRngString[i] = alphanumericCharset[random.Intn(len(alphanumericCharset))]
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

	// Create a buffered reader from the torrent file
	torrent := Torrent{
		OurPeerId:            ourPeerId,
		Bitfield:             bitfield,
		InfoHashHex:          infoHash,
		InfoHash:             string(infoHash[:]),
		TorrentFile:          *torrentFile,
		PieceProgresses:      make([]PieceProgress, len(pieceHashes)),
		Status:               TorrentStatusPaused,
		DownloadedPieceCount: 0,
		TotalPieceCount:      uint32(len(pieceHashes)),
		Peers:                make(map[string]peer.Peer),
	}

	pieceProgresses := make([]PieceProgress, len(pieceHashes))
	for i, hash := range pieceHashes {
		length := torrent.calculatePieceSize(uint32(i))
		pieceProgresses = append(pieceProgresses, PieceProgress{
			Buffer:              make([]byte, 0, length),
			Index:               uint32(i),
			DownloadedByteCount: 0,
			RequestedByteCount:  0,
			Length:              length,
			Hash:                hash,
		})
	}

	return &torrent, nil
}

// buildTrackerUrl builds a tracker URL for the Torrent.
//
// It takes a trackerAddress string and ourPort uint16 as parameters and returns a *url.URL and an error.
func (t *Torrent) buildTrackerUrl(trackerAddress string, ourPort uint16) (*url.URL, error) {
	trackerUrl, err := url.Parse(trackerAddress)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(t.OurPeerId[:])},
		"port":       []string{strconv.Itoa(int(ourPort))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"0"},
		"left":       []string{strconv.Itoa(int(t.Length))},
	}

	trackerUrl.RawQuery = params.Encode()

	return trackerUrl, nil
}

func (t *Torrent) DownloadTorrent() error {
	trackerUrl, err := t.buildTrackerUrl(t.TorrentFile.Announce, 6881)
	if err != nil {
		return err
	}

	tracker, err := tracker.New(trackerUrl)
	if err != nil {
		return err
	}

	t.createPeers(tracker.Peers)

	t.createPeerWorkers()

	return nil
}

func (t *Torrent) createPeers(peerAddresses []tracker.PeerAddress) {
	for _, peerAddress := range peerAddresses {
		fullPeerAddress := peerAddress.Ip + ":" + strconv.Itoa(int(peerAddress.Port))

		t.Peers[fullPeerAddress] = peer.New(peerAddress, fullPeerAddress, peer.PeerStatusDisconnected, "")
	}
}

func (t *Torrent) calculateBoundsForPiece(index uint32) (begin uint32, end uint32) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index uint32) uint32 {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) createPeerWorkers() {
	pieceWorkQueue := make(chan *PieceProgress, t.TotalPieceCount)
	results := make(chan *PieceProgress)

	for _, pieceProgress := range t.PieceProgresses {
		pieceWorkQueue <- &pieceProgress
	}

	for _, peer := range t.Peers {
		go t.startPeerWorker(peer, pieceWorkQueue, results)
	}

	//open file for writing pieces
	file, err := os.OpenFile(fmt.Sprintf("./%s", t.TorrentFile.Info.Name), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	donePieces := 0

	for donePieces < int(t.TotalPieceCount) {
		result := <-results
		begin, _ := t.calculateBoundsForPiece(result.Index)

		// Seek to the beginning index
		_, err := file.Seek(int64(begin), 0)
		if err != nil {
			log.Fatal(err)
		}

		// Write piece to the specified range
		_, err = file.Write(result.Buffer) // Replace with the data you want to write
		if err != nil {
			log.Fatal(err)
		}

		donePieces++

		percent := float64(donePieces) / float64(t.TotalPieceCount) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, result.Index, numWorkers)
	}
}

func (t *Torrent) startPeerWorker(peer peer.Peer, pieceWorks chan *PieceProgress, results chan *PieceProgress) {
	peerClient, err := peer.NewClient(
		t.InfoHash,
		int(t.TotalPieceCount),
		t.OurPeerId,
		t.Bitfield,
	)
	if err != nil {
		return
	}
}
