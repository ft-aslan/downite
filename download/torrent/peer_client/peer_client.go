package peerclient

import (
	"downite/download/torrent/handshake"
	"net"
	"time"
)

type PeerClient struct {
	tcp_stream net.Conn
	choked     bool
	peer       Peer
	bitfield   []byte
}

func new(
	peer Peer,
	infoHash []byte,
	totalPieceCount int,
	ourPeerId string,
) (*PeerClient, error) {

	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	handshake := handshake.New(infoHash, ourPeerId)
	handshakeResult, err =
		handshakeWithPeer(conn, address, &handshake.serialize())

	if err != nil {
		return nil, err
	}

	return &PeerClient{
		tcp_stream,
		peer,
		choked:   true,
		bitfield: make([]byte, 0, totalPieceCount),
	}
}

func handshakeWithPeer(
	conn *net.Conn,
	address string,
	handshakeString []byte,
) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	_, err := conn.Write(handshakeString)
	if err != nil {
		return nil, err
	}

	h, err := handshake.Read(conn)

}
