package handshake

import (
	"fmt"
	"io"
)

/*
HANDSHAKE IS <pstrlen><pstr><reserved><info_hash><peer_id>

	1 byte:	19 byte: 8 byte:  20 byte : 20 byte  = total length is pstr+49

pstr is string identifier of the protocol and it may not be BitTorrent protocol and it may not be 19 byte. So check for pstrlen. Its the length of pstr
*/
type Handshake struct {
	//pstr is string identifier of the protocol
	pstr     []byte
	infoHash [20]byte
	peerId   [20]byte
}

func New(infoHash [20]byte, ourPeerId [20]byte) *Handshake {
	return &Handshake{
		pstr:     []byte("BitTorrent protocol"),
		infoHash: infoHash,
		peerId:   ourPeerId,
	}
}

func (h *Handshake) serialize() []byte {
	buffer := make([]byte, len(h.pstr)+49)
	buffer[0] = byte(len(h.pstr))
	currentIndex := 1
	currentIndex += copy(buffer[currentIndex:], h.pstr)
	currentIndex += copy(buffer[currentIndex:], make([]byte, 8)) // 8 reserved bytes
	currentIndex += copy(buffer[currentIndex:], h.infoHash[:])
	currentIndex += copy(buffer[currentIndex:], h.peerId[:])
	return buffer
}

// Receive handshake message
func Read(reader io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(reader, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}
	handshakeBuffer := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(reader, handshakeBuffer)
	if err != nil {
		return nil, err
	}

	// Verify handshake message
	// pstr is BitTorrent protocol
	// pstrlen is length of ptstr. its always 19 bytes
	pstr := handshakeBuffer[1 : pstrlen+1]
	if string(pstr) != "BitTorrent protocol" {
		err := fmt.Errorf("peer is not using BitTorrent protocol")
		return nil, err
	}
	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuffer[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuffer[pstrlen+8+20:])

	h := Handshake{
		pstr:     pstr,
		infoHash: infoHash,
		peerId:   peerID,
	}

	return &h, nil
}
