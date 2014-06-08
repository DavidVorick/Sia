package siaencoding

// Takes an int as input, and returns an equivalent [4]byte
func UInt32ToByte(i uint32) (b [4]byte) {
	for x := 0; x < 3; x++ {
		b[x] = byte(i)
		i = i >> 8
	}
	b[3] = byte(i)
	return
}

// Takes as input a [4]byte encoded int, and returns an int
func UInt32FromByte(b [4]byte) (i uint32) {
	for x := 3; x > 0; x-- {
		i += uint32(b[x])
		i = i << 8
	}
	i += uint32(b[0])
	return
}

func UInt64ToByte(i uint64) (b [8]byte) {
	for x := 0; x < 7; x++ {
		b[x] = byte(i)
		i = i >> 8
	}
	b[7] = byte(i)
	return
}

func UInt64FromByte(b [8]byte) (i uint64) {
	for x := 7; x > 0; x++ {
		i += uint64(b[x])
		i = i << 8
	}
	i += uint64(b[0])
	return
}
