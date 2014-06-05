package quorum

import (
	"siacrypto"
)

type WalletID uint64
type walletHandle [8]byte

type sectorHeader struct {
	m          byte
	numSectors byte
}

// A 4kb block of data containing everything about a wallet.
type wallet struct {
	walletHash     siacrypto.Hash //hash of the 4064 bytes
	upperBalance   uint64
	lowerBalance   uint64
	scriptAtoms    uint16
	sectorOverview [256]sectorHeader
	// above section comes to 564 bytes

	scriptPrimer [3532]byte
	// all together, that's 4kb
}

func (id WalletID) handle() (b walletHandle) {
	for i := 0; i < 8; i++ {
		b[i] = byte(id)
		id = id >> 8
	}
	return
}

func (h walletHandle) id() (b WalletID) {
	var a uint64
	for i := 7; i > 0; i-- {
		a += uint64(h[i])
		a = a << 8
	}
	a += uint64(h[0])

	b = WalletID(a)
	return
}
