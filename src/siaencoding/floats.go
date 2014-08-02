package siaencoding

import (
	"encoding/binary"
	"unsafe"
)

// we don't need no math package

func float32bits(f float32) uint32 { return *(*uint32)(unsafe.Pointer(&f)) }

func float32frombits(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }

func float64bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }

func float64frombits(u uint64) float64 { return *(*float64)(unsafe.Pointer(&u)) }

// EncFloat32 encodes a float32 as a slice of 4 bytes.
func EncFloat32(f float32) (b []byte) {
	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, float32bits(f))
	return
}

// DecFloat32 decodes a slice of 4 bytes into a float32.
// It panics if len(b) < 4.
func DecFloat32(b []byte) (f float32) {
	f = float32frombits(binary.LittleEndian.Uint32(b))
	return
}

// EncFloat64 encodes a float64 as a slice of 8 bytes.
func EncFloat64(f float64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, float64bits(f))
	return
}

// DecFloat64 decodes a slice of 8 bytes into a float64.
// It panics if len(b) < 8.
func DecFloat64(b []byte) (f float64) {
	f = float64frombits(binary.LittleEndian.Uint64(b))
	return
}
