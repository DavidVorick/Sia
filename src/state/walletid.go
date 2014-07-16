package state

import (
	"siaencoding"
)

type WalletID uint64

func (id WalletID) Bytes() []byte {
	return siaencoding.EncUint64(uint64(id))
}
