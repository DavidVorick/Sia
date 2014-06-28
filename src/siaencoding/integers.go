package siaencoding

import (
	"encoding/binary"
	"math/big"
)

// really just a bytewise reversal
func ToggleEndianness(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)

	i, j := 0, len(b)-1
	for i < j {
		r[i], r[j] = r[j], r[i]
		i, j = i+1, j-1
	}
	return r
}

func EncUint16(i uint16) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return
}

func DecUint16(b []byte) (i uint16) {
	i = binary.LittleEndian.Uint16(b)
	return
}

func EncUint32(i uint32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return
}

func DecUint32(b []byte) (i uint32) {
	i = binary.LittleEndian.Uint32(b)
	return
}

func EncUint64(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return
}

func DecUint64(b []byte) (i uint64) {
	i = binary.LittleEndian.Uint64(b)
	return
}

func EncInt32(i int32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return
}

func DecInt32(b []byte) (i int32) {
	i = int32(binary.LittleEndian.Uint32(b))
	return
}

func EncInt64(i int64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return
}

func DecInt64(b []byte) (i int64) {
	i = int64(binary.LittleEndian.Uint64(b))
	return
}

func EncUint128(i *big.Int) (b []byte) {
	b = make([]byte, 16)
	copy(b, ToggleEndianness(i.Bytes()))
	return
}

func DecUint128(b []byte) (i *big.Int) {
	i = new(big.Int)
	i.SetBytes(ToggleEndianness(b))
	return
}
