package quorum

type sectorHeader struct {
	crc   [6]byte
	m     byte
	atoms byte
}
