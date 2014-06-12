package quorum

import (
	"siaencoding"
)

const (
	WalletIDSize = 8
)

type WalletID uint64

func (id WalletID) Bytes() [8]byte {
	return siaencoding.UInt64ToByte(uint64(id))
}
