package siaencoding

import (
	"encoding/binary"
)

func UInt32ToByte(i uint32) (b [4]byte) {
	binary.LittleEndian.PutUint32(b[:], i)
	return
}

func UInt32FromByte(b []byte) (i uint32) {
	i = binary.LittleEndian.Uint32(b)
	return
}

func UInt64ToByte(i uint64) (b [8]byte) {
	binary.LittleEndian.PutUint64(b[:], i)
	return
}

func UInt64FromByte(b []byte) (i uint64) {
	i = binary.LittleEndian.Uint64(b)
	return
}
