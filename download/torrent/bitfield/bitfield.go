package bitfield

type Bitfield []byte

func (bf Bitfield) GetPiece(index uint32) bool {
	byteIndex := index / 8
	binIndex := index % 8
	return bf[byteIndex]&(1<<uint(7-binIndex)) != 0
}
func (bf Bitfield) SetPiece(index uint32) {
	byteIndex := index / 8
	binIndex := index % 8
	bf[byteIndex] |= 1 << uint(7-binIndex)
}
