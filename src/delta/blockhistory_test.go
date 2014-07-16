package delta

import (
	"testing"
)

// TestSaveAndLoadBlock saves SnapshotLength+2 blocks, and after saving each
// block it tries to load every single block. Right now TestSaveAndLoadBlock
// does not check to see that recentHistory gets deleted.
func TestSaveAndLoadBlock(t *testing.T) {
	e := Engine{
		filePrefix:          "../../filesCreatedDuringTesting/TestSaveAndLoad",
		activeHistoryHead:   ^uint32(0) - uint32(SnapshotLength-1),
		activeHistoryLength: SnapshotLength,
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
