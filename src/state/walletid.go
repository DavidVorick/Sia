package state

import (
	"fmt"
	"siaencoding"
)

const (
	WalletIDSize = 8
)

type WalletID uint64

func (id WalletID) Bytes() []byte {
	// replace with siaencoding.EncWalletID(id) ?
	return siaencoding.EncUint64(uint64(id))
}

func (id *WalletID) GobEncode() (gobID []byte, err error) {
	if id == nil {
		err = fmt.Errorf("Cannot encode nil id")
		return
	}

	gobID = siaencoding.EncUint64(uint64(*id))
	return
}

func (id *WalletID) GobDecode(gobID []byte) (err error) {
	if id == nil {
		err = fmt.Errorf("Cannot decode into nil id")
		return
	}
	if len(gobID) != WalletIDSize {
		err = fmt.Errorf("gobID is of incorrect size")
		return
	}

	*id = WalletID(siaencoding.DecUint64(gobID))
	return
}
