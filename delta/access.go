package delta

import (
	"github.com/NebulousLabs/Sia/state"
)

func (e *Engine) BuildStorageProof() (sp state.StorageProof, err error) {
	return e.state.BuildStorageProof()
}

func (e *Engine) SegmentFilename(id state.WalletID) string {
	return e.state.SectorFilename(id)
}

func (e *Engine) RepairChan() chan state.WalletID {
	return e.state.RepairChan
}
