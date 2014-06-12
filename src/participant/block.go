package participant

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siaencoding"
)

const (
	SnapshotLen            = 3 // number of blocks separating each snapshot
	BlockHistoryHeaderSize = 4 + SnapshotLen*4 + siacrypto.TruncatedHashSize*SnapshotLen
)

// A block is just a collection of heartbeats, along with information about
// which block it is and which block was it's parent.
type block struct {
	height     uint32
	parent     siacrypto.TruncatedHash
	heartbeats [quorum.QuorumSize]*heartbeat
}

// the blockHistoryHeader is the header that preceeds the block history file,
// containing an index of all the blocks in the history and their data offsets
// in the file.
type blockHistoryHeader struct {
	latestBlock  uint32
	blockOffsets [SnapshotLen]uint32
}

func (bhh *blockHistoryHeader) GobEncode() (gobBHH []byte, err error) {
	gobBHH = make([]byte, BlockHistoryHeaderSize)
	encodedInt := siaencoding.EncUint32(bhh.latestBlock)
	copy(gobBHH, encodedInt[:])
	offset := len(encodedInt)

	for i := range bhh.blockOffsets {
		encodedInt = siaencoding.EncUint32(bhh.blockOffsets[i])
		copy(gobBHH[offset:], encodedInt[:])
		offset += len(encodedInt)
	}

	return
}

func (bhh *blockHistoryHeader) GobDecode(gobBHH []byte) (err error) {
	if len(gobBHH) != BlockHistoryHeaderSize {
		err = fmt.Errorf("gobBHH has wrong size, cannot decode!")
	}

	bhh.latestBlock = siaencoding.DecUint32(gobBHH[:4])
	offset := 4

	for i := range bhh.blockOffsets {
		bhh.blockOffsets[i] = siaencoding.DecUint32(gobBHH[offset : offset+4])
		offset += 4
	}

	return
}

// Takes a hash of the height, parent, and heartbeats field of a block
func (b *block) Hash() (hash siacrypto.TruncatedHash, err error) {
	if b == nil {
		err = fmt.Errorf("Cannot hash a nil block")
		return
	}

	gobBlock, err := b.GobEncode()
	if err != nil {
		return
	}
	hash, err = siacrypto.CalculateTruncatedHash(gobBlock)
	return
}

func (b *block) GobEncode() (gobBlock []byte, err error) {
	// get the height as a [4]byte
	intb := siaencoding.EncUint32(b.height)

	// get all of the heartbeats in their encoded form
	var encodedHeartbeats [quorum.QuorumSize][]byte
	for i, heartbeat := range b.heartbeats {
		if heartbeat == nil {
			encodedHeartbeats[i] = nil
		} else {
			var encodedHeartbeat []byte
			encodedHeartbeat, err = heartbeat.GobEncode()
			if err != nil {
				return
			}
			encodedHeartbeats[i] = encodedHeartbeat
		}
	}

	// calculate total size
	size := len(intb)     // height
	size += len(b.parent) // parent hash
	for i := range encodedHeartbeats {
		size += len(encodedHeartbeats[i]) // all the heartbeats
	}
	size += len(intb) * quorum.QuorumSize // an offset for each heartbeat

	// encode height, then parent, then heartbeats
	gobBlock = make([]byte, size)
	copy(gobBlock, intb[:])
	offset := len(intb)
	copy(gobBlock[offset:], b.parent[:])
	offset += siacrypto.TruncatedHashSize

	// get the offset for the first heartbeat
	heartbeatOffset := offset + len(intb)*quorum.QuorumSize
	for i := range encodedHeartbeats {
		// encode nil heartbeats as -1, or all 1's for uint
		if encodedHeartbeats[i] == nil {
			intb = siaencoding.EncUint32(^uint32(0))
			copy(gobBlock[offset:], intb[:])
		} else {
			intb = siaencoding.EncUint32(uint32(heartbeatOffset))
			copy(gobBlock[offset:], intb[:])
			copy(gobBlock[heartbeatOffset:], encodedHeartbeats[i])
			heartbeatOffset += len(encodedHeartbeats[i])
		}
		offset += len(intb)
	}

	return
}

func (b *block) GobDecode(gobBlock []byte) (err error) {
	if b == nil {
		err = fmt.Errorf("cannot decode into nil value")
		return
	}

	// minimum size for a block is the height, parent hash, and offsets for each
	// of quorum.QuorumSize heartbeats
	if len(gobBlock) < siacrypto.TruncatedHashSize+quorum.QuorumSize*4+4 {
		err = fmt.Errorf("invalid gob block")
		return
	}

	// decode height and parent
	b.height = siaencoding.DecUint32(gobBlock[:4])
	offset := 4
	copy(b.parent[:], gobBlock[offset:offset+siacrypto.TruncatedHashSize])
	offset += siacrypto.TruncatedHashSize

	// decode each heartbeat
	var nextOffset uint32
	var heartbeatOffset uint32
	var i int
	for i = 0; i < quorum.QuorumSize-1; i++ {
		// get the offset for the current heartbeat
		heartbeatOffset = siaencoding.DecUint32(gobBlock[offset : offset+4])
		offset += 4
		if heartbeatOffset == ^uint32(0) {
			b.heartbeats[i] = nil
			continue
		}

		// get the offset for the next heartbeat (to know the length of this
		// heartbeat)
		nextOffset = siaencoding.DecUint32(gobBlock[offset : offset+4])

		// in the loop, the +1 is derived from the fact that offset has already
		// been advanced after 'i'
		j := 1
		for nextOffset == ^uint32(0) && j+i+1 < quorum.QuorumSize {
			nextOffset = siaencoding.DecUint32(gobBlock[offset+4*j : offset+4*j+4])
			j++
		}

		if nextOffset == ^uint32(0) {
			// current heartbeat is last heartbeat
			break
		}

		if heartbeatOffset > nextOffset || int(nextOffset)+MinHeartbeatSize > len(gobBlock) {
			err = fmt.Errorf("got invalid gob block")
			return
		}

		b.heartbeats[i] = new(heartbeat)
		err = b.heartbeats[i].GobDecode(gobBlock[heartbeatOffset:nextOffset])
		if err != nil {
			return
		}
	}

	// if nextOffset is nil, then the program broke because the last heartbeat
	// has been loaded. If nextOffset is not nil, then the program broke because
	// the for loop expired, but there's still a heartbeat dangling at the end
	b.heartbeats[i] = new(heartbeat)
	if nextOffset != ^uint32(0) {
		heartbeatOffset = siaencoding.DecUint32(gobBlock[offset : offset+4])
	}
	b.heartbeats[i].GobDecode(gobBlock[heartbeatOffset:])

	return
}

// SaveBlock takes a block, which is just a list of heartbeats attatched to
// non-tossed participants, and saves it to disk. Blocks are saved in chains of
// SnapshotLen, after which a new chain is created and the oldest is deleted.
// Two chains total are kept around, one complete chain and one yet-incomplete
// chain. These chains are purely so other hosts have large windows to download
// the quorum and synchronize with their siblings.
func (p *Participant) saveBlock(b *block) (err error) {
	// if p.activHistoryStep == SnapshotLen, it's time to cycle the histories by
	// deleting the oldest one and create a new one. Otherwise you just append to
	// the existing and yet-incomplete history.
	var file *os.File
	if p.activeHistoryStep == SnapshotLen {
		// reset history step, delete old history, migrate recently-completed
		// history.
		p.quorum.SaveSnap()
		p.activeHistoryStep = 0
		os.Remove(p.recentHistory)
		p.recentHistory = p.activeHistory

		// find a name for the new history and create a file for it
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
			println(n)
			panic(err)
		}

		err = bs[i].GobDecode(blockBytes)
		if err != nil {
			panic(err)
		}
	}

	return
}
