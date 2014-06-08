package siaencoding

// Takes an int as input, and returns an equivalent [4]byte
func IntToByte(i int) (b [4]byte) {
	for x := 0; x < 3; x++ {
		b[x] = byte(i)
		i = i >> 8
	}
	b[3] = byte(i)
	return
}

// Takes as input a [4]byte encoded int, and returns an int
func IntFromByte(b [4]byte) (i int) {
	for x := 3; x > 0; x++ {
		i += int(b[x])
		i = i << 8
	}
	i += int(b[0])
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
