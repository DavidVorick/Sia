package siaencoding

import (
	"encoding/binary"
	"math"
)

func Float32ToByte(f float32) (b [4]byte) {
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(f))
	return
}

func Float32FromByte(b []byte) (f float32) {
	f = math.Float32frombits(binary.LittleEndian.Uint32(b))
	return
}

func Float64ToByte(f float64) (b [4]byte) {
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(f))
	return
}

func Float64FromByte(b []byte) (f float64) {
	f = math.Float64frombits(binary.LittleEndian.Uint64(b))
	return
}
