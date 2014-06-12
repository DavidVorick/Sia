package quorum

import (
	"siaencoding"
)

const (
	WalletIDSize = 8
)

type WalletID uint64

func (id WalletID) Bytes() [8]byte {
	// replace with siaencoding.EncWalletID(id) ?
	return siaencoding.EncUint64(uint64(id))
}
