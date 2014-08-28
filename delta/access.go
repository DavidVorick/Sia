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
