package delta

import (
	"fmt"
	"os"
)

// historyFilename is a helper function used to unify recentHistoryFilename and
// activeHistoryFilename, and should only be called by these two functions.
// Other history filenames are not guaranteed to exist.
func (e *Engine) historyFilename(head uint32) string {
	return fmt.Sprintf("%s.snapshot.%v", e.filePrefix, head)
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
	return e.historyFilename(e.recentHistoryHead + SnapshotLength)
}

// SaveBlock takes a block and saves it to disk. Blocks are saved in chains of
// SnapshotLen, after which a new chain is created and the oldest is deleted.
// Two chains total are kept, one complete chain and one yet-incomplete chain.
// Two chains are kept so that hosts synchronizing to the network can have
// large windows of time to download the quorum and blocks before they risk
// downloading obsolete data.
func (e *Engine) saveBlock(b *block) (err error) {
	// if e.activeHistoryLen == SnapshotLen, the old complete history is deleted
	// and replaced by the activeHistory. Then the activeHistory is replaced by a
	// new history which will start with a single block and be of length 1.
	var file *os.File
	if e.activeHistoryLength == SnapshotLength {
		// reset history step, delete old history, migrate recently-completed
		// history.
		e.SaveSnapshot()
		os.Remove(e.recentHistoryFilename())
		e.activeHistoryLength = 0
		e.recentHistoryHead += SnapshotLength

		// create a new activeHistory file
		p.activeHistory = p.quorum.GetWalletPrefix()
		p.activeHistory += fmt.Sprintf("blockHistory.%v", b.height)
		file, err = os.Create(p.activeHistory)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		// increase the active step and open the existing file for writing.
		file, err = os.OpenFile(p.activeHistory, os.O_RDWR, 0666)
		if err != nil {
			panic(p.activeHistory)
		}
		defer file.Close()
	}

	// if p.activeHistoryStep == 0, there is nothing to load. Otherwise, the saved
	// blockHistoryHeader must be loaded from the file into memory.
	var bhh blockHistoryHeader
	if p.activeHistoryStep != 0 {
		blockHistoryHeaderBytes := make([]byte, BlockHistoryHeaderSize)
		n, err := file.Read(blockHistoryHeaderBytes)
		if err != nil || n != BlockHistoryHeaderSize {
			panic(err)
		}
		err = bhh.GobDecode(blockHistoryHeaderBytes)
		if err != nil {
			panic(err)
		}
	}

	// encode the block so it can be saved to disk
	gobBlock, err := b.GobEncode()
	if err != nil {
		return
	}

	// if the latest block is 0, then the struct is empty and the first index
	// needs to account for the bhh struct. If the latest block is SnapshotLen-1,
	// then it is the last block and updating the following index will cause a
	// panic
	if bhh.latestBlock == 0 {
		bhh.blockOffsets[0] = uint32(BlockHistoryHeaderSize)
	}
	if bhh.latestBlock != SnapshotLen-1 {
		bhh.blockOffsets[bhh.latestBlock+1] += uint32(len(gobBlock)) + bhh.blockOffsets[bhh.latestBlock]
	}
	bhh.latestBlock += 1

	// seek back to 0 to write the updated bhh struct to disk
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	encodeBHH, err := bhh.GobEncode()
	if err != nil {
		panic(err)
	}
	n, err := file.Write(encodeBHH)
	if err != nil || n != len(encodeBHH) {
		panic(err)
	}

	// seek to the offset location in the file to write the block to disk
	_, err = file.Seek(int64(bhh.blockOffsets[bhh.latestBlock-1]), 0)
	if err != nil {
		panic(err)
	}
	n, err = file.Write(gobBlock)
	if err != nil || n != len(gobBlock) {
		panic(err)
	}
	p.activeHistoryStep += 1

	return
}

func (p *Participant) loadBlocks(snapshot bool) (bs []block) {
	var file *os.File
	var err error
	if snapshot == p.quorum.CurrentSnapshot() {
		file, err = os.Open(p.activeHistory)
	} else {
		file, err = os.Open(p.recentHistory)
	}
	if err != nil {
		panic(err)
	}

	var bhh blockHistoryHeader
	bhhBytes := make([]byte, BlockHistoryHeaderSize)
	n, err := file.Read(bhhBytes)
	if err != nil || n != BlockHistoryHeaderSize {
		panic(err)
	}
	err = bhh.GobDecode(bhhBytes)
	if err != nil {
		panic(err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	bs = make([]block, bhh.latestBlock)
	for i := uint32(0); i < bhh.latestBlock; i++ {
		var byteCount uint32
		if i == SnapshotLen-1 {
			byteCount = uint32(fileSize) - bhh.blockOffsets[i]
		} else {
			byteCount = bhh.blockOffsets[i+1] - bhh.blockOffsets[i]
		}
		blockBytes := make([]byte, byteCount)

		n, err = file.Read(blockBytes)
		if err != nil || n != int(byteCount) {
			panic(err)
		}

		err = bs[i].GobDecode(blockBytes)
		if err != nil {
			panic(err)
		}
	}

	return
}
