package quorum

import (
	"encoding/base64"
	"fmt"
	"os"
	"siacrypto"
	"strings"
)

const (
	scriptPrimerSize = 3532
)

type WalletID uint64
type walletHandle [8]byte

type sectorHeader struct {
	m        byte
	numAtoms byte
}

// A 4kb block of data containing everything about a wallet.
type wallet struct {
	walletHash     siacrypto.TruncatedHash //hash of the 4064 bytes
	upperBalance   uint64
	lowerBalance   uint64
	scriptAtoms    uint16
	sectorOverview [256]sectorHeader
	// above section comes to 564 bytes

	scriptPrimer [scriptPrimerSize]byte
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

func (w *wallet) Bytes() (b [4096]byte) {
	copy(b[:], w.walletHash[:])
	offset := siacrypto.TruncatedHashSize

	tmp := w.upperBalance
	for i := 0; i < 8; i++ {
		b[offset+i] = byte(tmp)
		tmp = tmp >> 8
	}
	offset += 8

	tmp = w.lowerBalance
	for i := 0; i < 8; i++ {
		b[offset+i] = byte(tmp)
		tmp = tmp >> 8
	}
	offset += 8

	tmp16 := w.scriptAtoms
	for i := 0; i < 2; i++ {
		b[offset+i] = byte(tmp16)
		tmp = tmp >> 8
	}
	offset += 2

	for i, sector := range w.sectorOverview {
		b[offset+i*2] = sector.m
		b[offset+i*2+1] = sector.numAtoms
	}
	offset += 2 * len(w.sectorOverview)

	copy(b[offset:], w.scriptPrimer[:])
	return
}
