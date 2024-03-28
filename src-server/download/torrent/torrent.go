package torrent

import (
	"bytes"
	"crypto/sha1"
	"downite/download/torrent/bitfield"
	"downite/download/torrent/decoding"
	"downite/download/torrent/message"
	"downite/download/torrent/peer"
	"downite/download/torrent/tracker"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"
)

const defaultPieceLength = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5
const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type TorrentStatus int

const (
	TorrentStatusPaused TorrentStatus = iota
	TorrentStatusCompleted
	TorrentStatusDownloading
)

type PieceProgress struct {
	Index               int
	Buffer              []byte
	DownloadedByteCount int
	RequestedByteCount  int
	Backlog             int
	Length              int
	Hash                [20]byte // pub requested: u32,
}

type Torrent struct {
	id                   string
	OurPeerId            [20]byte
	TorrentFile          decoding.TorrentFile
	InfoHash             [20]byte // hash of info field.
	Bitfield             bitfield.Bitfield
	PieceProgresses      []PieceProgress
	Status               TorrentStatus
	DownloadedPieceCount int
	TotalPieceCount      int
	Length               int
	PieceLength          int
	Peers                map[string]peer.Peer
}

func New(torrentFilePath string) (*Torrent, error) {
	// Open the torrent rawFile
	raw_file, err := os.Open(torrentFilePath)
	if err != nil {
		return nil, err
	}
	defer raw_file.Close()

	var reader io.Reader = raw_file
	torrentFile, err := decoding.DecodeTorrentFile(reader)

	if err != nil {
		return nil, err
	}

	// Create our peer id
	//random generator for our peer id
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	peerIdHead := []byte("-DN0001-")
	// Create a random number generator
	peerIdRngString := make([]byte, 12)
	for i := range peerIdRngString {
		peerIdRngString[i] = alphanumericCharset[random.Intn(len(alphanumericCharset))]
	}

	ourPeerId := [20]byte{}
	copy(ourPeerId[:8], peerIdHead)
	copy(ourPeerId[8:], peerIdRngString)

	bitfield := make([]byte, 0, len(torrentFile.Info.Pieces)/8)

	//convert info field to sha1 hash
	infoHash, err := torrentFile.Info.Hash()
	if err != nil {
		return nil, err
	}

	//split pieces into 20 byte hashes
	pieceHashes, err := torrentFile.Info.SplitPieceHashes()
	if err != nil {
		return nil, err
	}

	// Create a buffered reader from the torrent file
	torrent := Torrent{
		OurPeerId:            ourPeerId,
		Bitfield:             bitfield,
		InfoHash:             infoHash,
		TorrentFile:          *torrentFile,
		PieceProgresses:      []PieceProgress{},
		Status:               TorrentStatusPaused,
		DownloadedPieceCount: 0,
		TotalPieceCount:      len(pieceHashes),
		Peers:                make(map[string]peer.Peer),
		PieceLength:          torrentFile.Info.PieceLength,
		Length:               torrentFile.Info.FileLength,
	}

	for i, hash := range pieceHashes {
		length := torrent.calculatePieceSize(i)
		torrent.PieceProgresses = append(torrent.PieceProgresses, PieceProgress{
			Buffer:              make([]byte, length),
			Index:               i,
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
		"compact":    []string{"1"},
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

func (t *Torrent) createPeers(peerAddresses []peer.PeerAddress) {
	for _, peerAddress := range peerAddresses {
		fullPeerAddress := net.JoinHostPort(peerAddress.Ip.String(), strconv.Itoa(int(peerAddress.Port)))

		t.Peers[fullPeerAddress] = peer.New(peerAddress, fullPeerAddress, peer.StatusDisconnected, "")
	}
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) createPeerWorkers() {
	pieceWorkQueue := make(chan *PieceProgress, t.TotalPieceCount)
	peerStatuses := make(chan *peer.Peer)
	results := make(chan *PieceProgress)

	for _, pieceProgress := range t.PieceProgresses {
		pieceWorkQueue <- &pieceProgress
	}

	for _, peer := range t.Peers {
		go t.startPeerWorker(peer, pieceWorkQueue, results)
	}

	file, err := os.OpenFile(fmt.Sprintf("./%s", t.TorrentFile.Info.Name), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// go t.watchPeers(peerStatuses)

	donePieces := 0
	// numWorkers := runtime.NumGoroutine() - 1
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
		t.Bitfield.SetPiece(result.Index)

		percent := float64(donePieces) / float64(t.TotalPieceCount) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, result.Index, numWorkers)
	}
	close(pieceWorkQueue)
	close(peerStatuses)
}
func (t *Torrent) watchPeers(peerStatues chan *peer.Peer) {
	for peer := range peerStatues {
		fmt.Printf("Peer: %s - Status: %s\n", peer.FullAddress, peer.Status)
	}
}

func (t *Torrent) startPeerWorker(peerNode peer.Peer, pieceWorks chan *PieceProgress, results chan *PieceProgress) {

	var peerClient *peer.PeerClient
	connectionTryCount := 3

	for len(pieceWorks) > 0 {
		var err error
		peerClient, err = peerNode.NewClient(
			t.InfoHash,
			t.TotalPieceCount,
			t.OurPeerId,
			t.Bitfield,
		)
		if err != nil {
			fmt.Println("Error connecting peer:", err)
			connectionTryCount--
			if connectionTryCount == 0 {
				peerNode.Status = peer.StatusDisconnected
				// peerStatuses <- &peerNode
				return
			}
			continue
		}
		break
	}
	defer peerClient.TcpConnection.Close()

	peerClient.SendMessage(message.NewMessage(message.IdUnchoke))
	peerClient.SendMessage(message.NewMessage(message.IdInterested))

	peerClient.TcpConnection.SetDeadline(time.Now().Add(30 * time.Second))
	defer peerClient.TcpConnection.SetDeadline(time.Time{}) // Disable the deadline

	for work := range pieceWorks {
		if !peerClient.Bitfield.GetPiece(work.Index) {
			pieceWorks <- work
			continue
		}

		for work.DownloadedByteCount < work.Length {
			if !peerClient.Choked {
				for work.Backlog < MaxBacklog && work.RequestedByteCount < work.Length {

					peerNode.Status = peer.StatusRequesting
					// peerStatuses <- &peerNode

					requestLength := defaultPieceLength

					leftByteCount := work.Length - work.RequestedByteCount
					if leftByteCount < defaultPieceLength {
						requestLength = leftByteCount
					}

					peerClient.SendMessage(
						message.NewRequestMessage(
							uint32(work.Index),
							uint32(work.RequestedByteCount),
							uint32(requestLength),
						),
					)
					work.RequestedByteCount += requestLength
					work.Backlog++

				}
			}

			// Read whatever message is available from client
			msg, err := peerClient.ReadMessage()
			if err != nil {
				// fmt.Println("Error reading message:", err)
				pieceWorks <- work
				continue
			}

			switch msg.Id {
			case message.IdUnchoke:
				peerClient.Choked = false
				peerNode.Status = peer.StatusUnchoked
				// peerStatuses <- &peerNode

			case message.IdChoke:
				peerClient.Choked = true
				peerNode.Status = peer.StatusChoked
				// peerStatuses <- &peerNode

			case message.IdHave:
				haveMsg, err := msg.ParseHaveMessage()
				if err != nil {
					// return err
					continue
				}
				peerClient.Bitfield.SetPiece(int(haveMsg.Index))
			case message.IdPiece:
				pieceMessage, err := msg.ParsePieceMessage()
				if err != nil {
					// return err
					continue
				}
				work.DownloadedByteCount += len(pieceMessage.Block)
				copy(work.Buffer[pieceMessage.Begin:], pieceMessage.Block)
				work.Backlog--

				peerNode.Status = peer.StatusDownloading
				// peerStatuses <- &peerNode
			}
		}
		err := checkPieceIntegrity(work)
		if err != nil {
			fmt.Println("Error piece integrity check:", err)
			continue
		}
		results <- work
		peerClient.SendMessage(message.NewHaveMessage(uint32(work.Index)))
	}
}

func checkPieceIntegrity(pieceProgress *PieceProgress) error {
	hash := sha1.Sum(pieceProgress.Buffer)
	if !bytes.Equal(hash[:], pieceProgress.Hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pieceProgress.Index)
	}
	return nil
}
