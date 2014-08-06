package siaencoding

import (
	"encoding/binary"
	"math"
)

// EncFloat32 encodes a float32 as a slice of 4 bytes.
func EncFloat32(f float32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(f))
	return
}

// DecFloat32 decodes a slice of 4 bytes into a float32.
// It panics if len(b) < 4.
func DecFloat32(b []byte) (f float32) {
	f = math.Float32frombits(binary.LittleEndian.Uint32(b))
	return
}

// EncFloat64 encodes a float64 as a slice of 8 bytes.
func EncFloat64(f float64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(f))
	return
}

// DecFloat64 decodes a slice of 8 bytes into a float64.
// It panics if len(b) < 8.
func DecFloat64(b []byte) (f float64) {
	f = math.Float64frombits(binary.LittleEndian.Uint64(b))
	return
}
