package delta

import (
	"fmt"
	"os"

	"github.com/NebulousLabs/Sia/siaencoding"
)

const (
	historyIndexLength = 4 * SnapshotLength
)

// historyFilename is a helper function used to unify recentHistoryFilename and
// activeHistoryFilename, and should only be called by these two functions.
// Other history filenames are not guaranteed to exist.
func (e *Engine) historyFilename(head uint32) string {
	return fmt.Sprintf("%sblockHistory.%v", e.filePrefix, head)
}

// recentHistoryFilename returns the name of the file containing the
// recentHistory, which is a set of SnapshotLength blocks connecting the least
// recent (but still maintained) snapshot to the most recently captured
// snapshot
func (e *Engine) recentHistoryFilename() string {
	return e.historyFilename(e.recentHistoryHead)
}

// activeHistoryFilename returns the name of the file containing the
// activeHistory, which is a set of blocks numbering less than or equal to
// SnapshotLength leading from the most recent snapshot to the current quorum.
func (e *Engine) activeHistoryFilename() string {
	return e.historyFilename(e.activeHistoryHead)
}

// saveBlock takes a block and saves it to disk. Blocks are saved in chains of
// SnapshotLen, after which a new chain is created and the oldest is deleted.
// Two chains total are kept, one complete chain and one yet-incomplete chain.
// Two chains are kept so that hosts synchronizing to the network can have
// large windows of time to download the quorum and blocks before they risk
// downloading obsolete data.
//
// The layout of the active history file is a series of encoded, 4 byte
// unsigned integers containing the byte-offsets of each block for the file.
// There are 'SnapshotLength' offsets, representing a 'historyIndexLength'
// prefix. Each points to the beginning of the next block. For convenience,
// each block is also prefixed with its own length.
func (e *Engine) saveBlock(b Block) (err error) {
	// if e.activeHistoryLen == SnapshotLen, the old complete history is deleted
	// and replaced by the activeHistory. Then the activeHistory is replaced by a
	// new history which will start with a single block and be of length 1.
	var file *os.File
	if e.activeHistoryLength == SnapshotLength {
		// remove the recent history file, and progress the recentHistoryHead
		os.Remove(e.recentHistoryFilename())
		e.recentHistoryHead = e.activeHistoryHead

		// reset activeHistoryLength, and progress the activeHistoryHead, then save
		// the snapshot. The ordering is important - the activeHistoryHead value
		// must be progressed before saveSnapshot() is called, such that the
		// snapshot is saved to the right filename.
		e.activeHistoryLength = 0
		e.activeHistoryHead += SnapshotLength
		err = e.saveSnapshot()
		if err != nil {
			return
		}

		// create a new activeHistory file
		file, err = os.Create(e.activeHistoryFilename())
		if err != nil {
			return
		}
		defer file.Close()
	} else {
		// increase the active step and open the existing file for writing.
		file, err = os.OpenFile(e.activeHistoryFilename(), os.O_RDWR, 0666)
		if err != nil {
			return
		}
		defer file.Close()
	}

	// encode the block to be saved in the block history
	encodedBlock, err := siaencoding.Marshal(b)
	if err != nil {
		return
	}

	// figure out the offset for this block
	var offset int
	if e.activeHistoryLength == 0 {
		offset = historyIndexLength

		// this is a special case where the previous save did not already set the
		// offset (because there was no previous save). We therefore need to play
		// that role and save the offset for the 0th block.
		encodedOffset := siaencoding.EncUint32(uint32(offset))
		_, err = file.Seek(0, 0) // These lines shoulnd't be needed
		if err != nil {          //
			return //
		} //
		file.Write(encodedOffset)
	} else {
		encodedOffset := make([]byte, 4)
		_, err = file.Seek(int64(e.activeHistoryLength*4), 0)
		if err != nil {
			return
		}
		_, err = file.Read(encodedOffset)
		if err != nil {
			return
		}
		offset = int(siaencoding.DecUint32(encodedOffset))
	}

	// save the offset of the next block, but only if there is a next block
	if e.activeHistoryLength != SnapshotLength-1 {
		nextOffset := offset + len(encodedBlock) + 4 // +4 for the length prefix
		_, err = file.Seek(int64(4+4*e.activeHistoryLength), 0)
		if err != nil {
			return
		}
		encodedNextOffset := siaencoding.EncUint32(uint32(nextOffset))
		_, err = file.Write(encodedNextOffset)
		if err != nil {
			return
		}
	}

	// Save a variable indicating the block size, followed by the block itself
	_, err = file.Seek(int64(offset), 0)
	if err != nil {
		return
	}
	blockSize := siaencoding.EncUint32(uint32(len(encodedBlock)))
	_, err = file.Write(blockSize)
	if err != nil {
		return
	}
	_, err = file.Write(encodedBlock)
	if err != nil {
		return
	}

	e.activeHistoryLength++
	return
}

// LoadBlock will check if the block in question is stored in one of the block
// history files, and then either return the block or an error.
func (e *Engine) LoadBlock(height uint32) (b Block, err error) {
	// Check for the block in active history and recent history, return an error
	// if it's not found in either location. Recent history may not exist yet, so
	// that possibility is also checked
	var file *os.File
	var blockIndex uint32
	if height >= e.activeHistoryHead && height < e.activeHistoryHead+e.activeHistoryLength {
		// block is in active history, load from that file
		file, err = os.Open(e.activeHistoryFilename())
		if err != nil {
			return
		}
		blockIndex = height - e.activeHistoryHead
	} else if e.recentHistoryHead != ^uint32(0) && height >= e.recentHistoryHead && height < e.recentHistoryHead+SnapshotLength {
		// block is in recent history, load from that file
		file, err = os.Open(e.recentHistoryFilename())
		if err != nil {
			return
		}
		blockIndex = height - e.recentHistoryHead
	} else {
		err = fmt.Errorf("LoadBlock: Block %v not in available history.", height)
		return
	}

	// Fetch the block from the determined index in the opened file.
	_, err = file.Seek(int64(4*blockIndex), 0)
	if err != nil {
		return
	}
	encodedBlockOffset := make([]byte, 4)
	_, err = file.Read(encodedBlockOffset)
	if err != nil {
		return
	}
	blockOffset := siaencoding.DecUint32(encodedBlockOffset)
	_, err = file.Seek(int64(blockOffset), 0)
	if err != nil {
		return
	}
	encodedBlockLength := make([]byte, 4)
	_, err = file.Read(encodedBlockLength)
	if err != nil {
		return
	}
	blockLength := siaencoding.DecUint32(encodedBlockLength)
	encodedBlock := make([]byte, blockLength)
	_, err = file.Read(encodedBlock)
	if err != nil {
		return
	}
	err = siaencoding.Unmarshal(encodedBlock, &b)
	return
}
