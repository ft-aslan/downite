package bitfield

type Bitfield []byte

func (bf Bitfield) GetPiece(index int) bool {
	byteIndex := index / 8
	binIndex := index % 8
	return bf[byteIndex]&(1<<uint(7-binIndex)) != 0
}

// SetPiece sets a bit in the bitfield
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// silently discard invalid bounded index
	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}
	bf[byteIndex] |= 1 << uint(7-offset)
}
