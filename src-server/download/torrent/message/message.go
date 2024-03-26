package message

import (
	"encoding/binary"
	"errors"
)

type MessageId uint8

const (
	IdChoke MessageId = iota
	IdUnchoke
	IdInterested
	IdNotInterested
	IdHave
	IdBitfield
	IdRequest
	IdPiece
	IdCancel
	IdPort
)

type Message struct {
	Id      MessageId
	Payload []byte
}
type PieceMessage struct {
	Index uint32
	Begin uint32
	Block []byte
}

type CancelMessage struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type PortMessage struct {
	Port uint16
}
type RequestMessage struct {
	Index  uint32
	Begin  uint32
	Length uint32
}
type BitfieldMessage struct {
	Bitfield []byte
}

func NewMessage(messageId MessageId) *Message {
	return &Message{
		Id: messageId,
	}
}
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	// <length prefix><message ID><payload>
	buffer := make([]byte, 4+1+len(m.Payload))
	binary.BigEndian.PutUint32(buffer[:4], uint32(1+len(m.Payload)))
	buffer[4] = byte(m.Id)
	copy(buffer[5:], m.Payload)
	return buffer
}

func NewBitfieldMessage(bitfield []byte) *Message {
	return &Message{
		Id:      IdBitfield,
		Payload: bitfield,
	}
}
func NewRequestMessage(index uint32, begin uint32, length uint32) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:], length)

	return &Message{
		Id:      IdRequest,
		Payload: payload,
	}
}
func NewPieceMessage(index uint32, begin uint32, block []byte) *Message {
	payload := make([]byte, 8+len(block))
	binary.BigEndian.PutUint32(payload[:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)

	copy(payload[8:], block)

	return &Message{
		Id:      IdPiece,
		Payload: payload,
	}
}
func NewCancelMessage(index uint32, begin uint32, length uint32) *Message {
	return NewRequestMessage(index, begin, length)
}
func NewHaveMessage(index uint32) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, index)
	return &Message{
		Id:      IdHave,
		Payload: payload,
	}
}
func NewPortMessage(port uint16) *Message {
	payload := make([]byte, 2)
	binary.BigEndian.PutUint16(payload, port)
	return &Message{
		Id:      IdPort,
		Payload: payload,
	}
}
func (m *Message) ParsePortMessage() (PortMessage, error) {
	if len(m.Payload) < 2 {
		return PortMessage{}, errors.New("invalid m.Payload length for PortMessage")
	}
	return PortMessage{
		Port: binary.BigEndian.Uint16(m.Payload),
	}, nil
}
func (m *Message) ParseBitfieldMessage() (BitfieldMessage, error) {
	if len(m.Payload) < 4 {
		return BitfieldMessage{}, errors.New("invalid m.Payload length for BitfieldMessage")
	}
	return BitfieldMessage{
		Bitfield: m.Payload[:],
	}, nil
}

// parsePieceMessage parses a PieceMessage from the given m.Payload.
func (m *Message) ParsePieceMessage() (PieceMessage, error) {
	if len(m.Payload) < 8 {
		return PieceMessage{}, errors.New("invalid m.Payload length for PieceMessage")
	}
	return PieceMessage{
		Index: binary.BigEndian.Uint32(m.Payload[:4]),
		Begin: binary.BigEndian.Uint32(m.Payload[4:8]),
		Block: m.Payload[8:],
	}, nil
}

// parseRequestMessage parses a RequestMessage from the given m.Payload.
func (m *Message) ParseRequestMessage() (RequestMessage, error) {
	if len(m.Payload) < 12 {
		return RequestMessage{}, errors.New("invalid m.Payload length for RequestMessage")
	}
	return RequestMessage{
		Index:  binary.BigEndian.Uint32(m.Payload[:4]),
		Begin:  binary.BigEndian.Uint32(m.Payload[4:8]),
		Length: binary.BigEndian.Uint32(m.Payload[8:]),
	}, nil
}

// parseCancelMessage parses a CancelMessage from the given m.Payload.
func (m *Message) ParseCancelMessage() (CancelMessage, error) {
	if len(m.Payload) < 12 {
		return CancelMessage{}, errors.New("invalid m.Payload length for CancelMessage")
	}
	return CancelMessage{
		Index:  binary.BigEndian.Uint32(m.Payload[:4]),
		Begin:  binary.BigEndian.Uint32(m.Payload[4:8]),
		Length: binary.BigEndian.Uint32(m.Payload[8:]),
	}, nil
}
