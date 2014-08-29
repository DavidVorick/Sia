package delta

import (
	"github.com/NebulousLabs/Sia/state"
)

func (e *Engine) BuildStorageProof() (sp state.StorageProof) {
	id, index, err := e.state.ProofLocation()
	if err != nil {
		return
	}
	sp = e.state.BuildStorageProof(id, index)
	return
}

func (e *Engine) SegmentFilename(id state.WalletID) string {
	return e.state.SectorFilename(id)
}

func (e *Engine) RepairChan() chan state.WalletID {
	return e.state.RepairChan
}
