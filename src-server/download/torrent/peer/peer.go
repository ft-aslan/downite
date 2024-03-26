package peer

import (
	"downite/download/torrent/bitfield"
	"downite/download/torrent/handshake"
	"downite/download/torrent/message"
	"downite/download/torrent/tracker"
	"encoding/binary"
	"io"
	"net"
	"time"
)

type PeerStatus int

const (
	StatusConnecting   PeerStatus = iota // Peer is in the process of establishing a connection
	StatusHandshake                      // Handshake initiated, waiting for peer's handshake
	StatusBitfield                       // Handshake completed, waiting for peer's bitfield
	StatusChoked                         // Peer has choked the connection, no data exchange
	StatusInterested                     // Peer is interested in our data, waiting to unchoke
	StatusUnchoked                       // Peer has been unchoked, data exchange allowed
	StatusRequesting                     // Requesting pieces from peer
	StatusDownloading                    // Downloading data from peer
	StatusSeeding                        // Uploading data to peer
	StatusDisconnected                   // Connection has been terminated
)

type Peer struct {
	Address     tracker.PeerAddress
	FullAddress string
	Status      PeerStatus
	Country     string
}
type PeerClient struct {
	TcpConnection net.Conn
	choked        bool
	peer          Peer
	Bitfield      bitfield.Bitfield
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
	bitfield []byte,
) (*PeerClient, error) {

	tcpConnection, err := net.DialTimeout("tcp", peer.FullAddress, 5*time.Second)
	if err != nil {
		return nil, err
	}
	handshake := handshake.New(infoHash, ourPeerId)
	_, err =
		handshakeWithPeer(tcpConnection.(*net.TCPConn), handshake.Serialize())

	if err != nil {
		return nil, err
	}

	peerClient := &PeerClient{
		TcpConnection: tcpConnection,
		peer:          *peer,
		choked:        true,
		Bitfield:      make([]byte, 0, totalPieceCount),
	}

	msg, err := peerClient.ReadMessage()
	if err != nil {
		return nil, err
	}

	bitfieldMessage, err := msg.ParseBitfieldMessage()
	if err != nil {
		return nil, err
	}
	peerClient.Bitfield = bitfieldMessage.Bitfield

	msg = message.NewBitfieldMessage(peerClient.Bitfield)
	_, err = tcpConnection.Write(msg.Serialize())

	if err != nil {
		return nil, err
	}
	return peerClient, nil
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

func (peer *PeerClient) SendMessage(message *message.Message) error {
	_, err := peer.TcpConnection.Write(message.Serialize())
	if err != nil {
		return err
	}
	return nil
}
func (peer *PeerClient) ReadMessage() (*message.Message, error) {

	// Read the length
	lengthBuffer := make([]byte, 4)
	_, err := io.ReadFull(peer.TcpConnection, lengthBuffer)
	length := binary.BigEndian.Uint32(lengthBuffer)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		//its keep-alive message
		return nil, err
	}

	// Read message ID
	messageIdBuffer := make([]byte, 1)
	_, err = io.ReadFull(peer.TcpConnection, messageIdBuffer)
	if err != nil {
		return nil, err
	}

	payload := make([]byte, length-1)
	if length > 1 {
		_, err = io.ReadFull(peer.TcpConnection, payload)
		if err != nil {
			return nil, err
		}
	}
	return &message.Message{
		Id:      message.MessageId(messageIdBuffer[0]),
		Payload: payload,
	}, nil
}
