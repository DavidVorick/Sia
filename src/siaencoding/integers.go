package siaencoding

import (
	"encoding/binary"
)

func EncUint32(i uint32) (b [4]byte) {
	binary.LittleEndian.PutUint32(b[:], i)
	return
}

func DecUint32(b []byte) (i uint32) {
	i = binary.LittleEndian.Uint32(b)
	return
}

func EncUint64(i uint64) (b [8]byte) {
	binary.LittleEndian.PutUint64(b[:], i)
	return
}

func DecUint64(b []byte) (i uint64) {
	i = binary.LittleEndian.Uint64(b)
	return
}

func EncInt32(i int32) (b [4]byte) {
	binary.LittleEndian.PutUint32(b[:], uint32(i))
	return
}

func DecInt32(b []byte) (i int32) {
	i = int32(binary.LittleEndian.Uint32(b))
	return
}

func EncInt64(i int64) (b [8]byte) {
	binary.LittleEndian.PutUint64(b[:], uint64(i))
	return
}

func DecInt64(b []byte) (i int64) {
	i = int64(binary.LittleEndian.Uint64(b))
	return
}
