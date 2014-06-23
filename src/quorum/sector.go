package quorum

func (q *Quorum) SectorFilename(id WalletID) (sectorFilename string) {
	walletFilename := q.walletFilename(id)
	sectorFilename = walletFilename + ".sector"
	return
}

func (q *Quorum) MerkleCollapse(filename string) {
	//
}
