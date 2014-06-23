package quorum

import (
	"siacrypto"
)

func (q *Quorum) SectorFilename(id WalletID) (sectorFilename string) {
	walletFilename := q.walletFilename(id)
	sectorFilename = walletFilename + ".sector"
	return
}

func (q *Quorum) MerkleCollapse(filename string) (hash siacrypto.Hash, err error) {
	return
}
