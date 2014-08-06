package siaencoding

import (
	"encoding/binary"
	"math/big"
)

// simple byte slice reversal
func toggleEndianness(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)

	i, j := 0, len(b)-1
	for i < j {
		r[i], r[j] = r[j], r[i]
		i, j = i+1, j-1
	}
	return r
}

// EncUint16 encodes a uint16 as a slice of 2 bytes.
func EncUint16(i uint16) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return
}

// DecUint16 decodes a slice of 2 bytes into a uint16.
// It panics if len(b) < 2.
func DecUint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

// EncUint32 encodes a uint32 as a slice of 4 bytes.
func EncUint32(i uint32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return
}

// DecUint32 decodes a slice of 4 bytes into a uint32.
// It panics if len(b) < 4.
func DecUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

// EncUint64 encodes a uint64 as a slice of 8 bytes.
func EncUint64(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return
}

// DecUint64 decodes a slice of 8 bytes into a uint64.
// It panics if len(b) < 8.
func DecUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// EncInt32 encodes an int32 as a slice of 4 bytes.
func EncInt32(i int32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return
}

// DecInt32 decodes a slice of 4 bytes into an int32.
// It panics if len(b) < 4.
func DecInt32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

// EncInt64 encodes an int64 as a slice of 8 bytes.
func EncInt64(i int64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return
}

// DecInt64 decodes a slice of 8 bytes into an int64.
// It panics if len(b) < 8.
func DecInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

// EncUint128 encodes a big.Int as a slice of 16 bytes.
// big.Ints are stored in big-endian format, so they must be converted
// to little-endian before being returned.
func EncUint128(i *big.Int) (b []byte) {
	b = make([]byte, 16)
	copy(b, toggleEndianness(i.Bytes()))
	return
}

// DecUint128 decodes a slice of 16 bytes into a big.Int
// It panics if len(b) < 16.
func DecUint128(b []byte) (i *big.Int) {
	i = new(big.Int)
	i.SetBytes(toggleEndianness(b))
	return
}
