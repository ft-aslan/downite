package peer

import (
	"downite/download/torrent/handshake"
	"downite/download/torrent/tracker"
	"net"
	"time"
)

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

type Peer struct {
	Address     tracker.PeerAddress
	FullAddress string
	Status      PeerStatus
	Country     string
}
type PeerClient struct {
	tcpConnection net.Conn
	choked        bool
	peer          Peer
	bitfield      []byte
}

func New(address tracker.PeerAddress, fullAddress string, status PeerStatus, country string) Peer {
	return Peer{
		Address:     address,
		FullAddress: fullAddress,
		Status:      status,
		Country:     country,
	}
}
func (peer *Peer) NewClient(
	infoHash [20]byte,
	totalPieceCount int,
	ourPeerId [20]byte,
) (*PeerClient, error) {

	tcpConnection, err := net.DialTimeout("tcp", peer.FullAddress, 3*time.Second)
	if err != nil {
		return nil, err
	}
	handshake := handshake.New(infoHash, ourPeerId)
	_, err =
		handshakeWithPeer(tcpConnection.(*net.TCPConn), handshake.Serialize())

	if err != nil {
		return nil, err
	}

	return &PeerClient{
		tcpConnection: tcpConnection,
		peer:          *peer,
		choked:        true,
		bitfield:      make([]byte, 0, totalPieceCount),
	}, nil
}

func handshakeWithPeer(
	conn *net.TCPConn,
	handshakeString []byte,
) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	_, err := conn.Write(handshakeString)
	if err != nil {
		return nil, err
	}

	h, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	return h, nil
}
