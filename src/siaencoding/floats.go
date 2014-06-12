package siaencoding

import (
	"encoding/binary"
	"math"
)

func EncFloat32(f float32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(f))
	return
}

func DecFloat32(b []byte) (f float32) {
	f = math.Float32frombits(binary.LittleEndian.Uint32(b))
	return
}

func EncFloat64(f float64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(f))
	return
}

func DecFloat64(b []byte) (f float64) {
	f = math.Float64frombits(binary.LittleEndian.Uint64(b))
	return
}
