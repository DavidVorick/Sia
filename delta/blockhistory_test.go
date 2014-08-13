package delta

import (
	"testing"

	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

// TestSaveAndLoadBlock saves SnapshotLength+2 blocks, and after saving each
// block it tries to load every single block. Right now TestSaveAndLoadBlock
// does not check to see that recentHistory gets deleted.
//
// Need to test that there are no recentHistoryHead errors regarding improper
// initialization.
func TestSaveAndLoadBlock(t *testing.T) {
	var e Engine
	e.SetFilePrefix(siafiles.TempFilename("TestSaveAndLoad"))
	e.SetSiblingIndex(0)
	err := e.Bootstrap(state.Sibling{
		WalletID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := uint32(0); i < SnapshotLength+2; i++ {
		b := Block{
			Height: i,
		}

		e.saveBlock(b)

		for j := uint32(0); j <= i; j++ {
			newBlock, err := e.LoadBlock(j)
			if err != nil {
				t.Error(err)
			}
			if newBlock.Height != j {
				t.Error("Failure upon loading a saved block")
			}
		}
	}
}
