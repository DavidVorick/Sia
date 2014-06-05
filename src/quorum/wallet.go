package quorum

import (
	"siacrypto"
)

type WalletID uint64

type sectorHeader struct {
	m byte
	numSectors byte
}

// A 4kb block of data containing everything about a wallet.
type wallet struct {
	walletHash siacrypto.Hash //hash of the 4064 bytes
	upperBalance uint64
	lowerBalance uint64
	scriptAtoms uint32
	sectorOverview [256]SectorHeader
	// above section comes to 564 bytes

	scriptPrimer [3532]byte
	// all together, that's 4kb
}
