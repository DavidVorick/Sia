package participant

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siaencoding"
)

const (
	SnapshotLen            = 100
	BlockHistoryHeaderSize = 4 + SnapshotLen*4 + siacrypto.TruncatedHashSize*SnapshotLen
)

type block struct {
	hash       siacrypto.TruncatedHash
	height     uint32
	parent     siacrypto.TruncatedHash
	heartbeats [quorum.QuorumSize]*heartbeat
}

type blockHistoryHeader struct {
	latestBlock  uint32
	blockOffsets [SnapshotLen]uint32
}

func (bhh *blockHistoryHeader) GobEncode() (gobBHH []byte, err error) {
	gobBHH = make([]byte, BlockHistoryHeaderSize)
	encodedInt := siaencoding.UInt32ToByte(bhh.latestBlock)
	copy(gobBHH, encodedInt[:])
	offset := len(encodedInt)

	for i := range bhh.blockOffsets {
		encodedInt = siaencoding.UInt32ToByte(bhh.blockOffsets[i])
		copy(gobBHH[offset:], encodedInt[:])
		offset += len(encodedInt)
	}

	return
}

func (bhh *blockHistoryHeader) GobDecode(gobBHH []byte) (err error) {
	if len(gobBHH) != BlockHistoryHeaderSize {
		err = fmt.Errorf("gobBHH has wrong size, cannot decode!")
	}

	var intb [4]byte
	copy(intb[:], gobBHH)
	bhh.latestBlock = siaencoding.UInt32FromByte(intb)
	offset := 4

	for i := range bhh.blockOffsets {
		copy(intb[:], gobBHH[offset:])
		bhh.blockOffsets[i] = siaencoding.UInt32FromByte(intb)
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

// The GobEncode for a block does not include the hash, that must be encoded
// separately.
func (b *block) GobEncode() (gobBlock []byte, err error) {
	intb := siaencoding.UInt32ToByte(b.height)
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
	heartbeatOffset := offset + len(intb)*quorum.QuorumSize
	for i := range encodedHeartbeats {
		if encodedHeartbeats[i] == nil {
			intb = siaencoding.UInt32ToByte(^uint32(0))
			copy(gobBlock[offset:], intb[:])
			offset += len(intb)
		} else {
			intb = siaencoding.UInt32ToByte(uint32(len(encodedHeartbeats[i])))
			copy(gobBlock[offset:], intb[:])
			copy(gobBlock[heartbeatOffset:], encodedHeartbeats[i])
			offset += len(intb)
			heartbeatOffset += len(encodedHeartbeats[i])
		}
	}

	return
}

// GobDecode for a block does not look for a hash
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
	var intb [4]byte
	copy(intb[:], gobBlock)
	b.height = siaencoding.UInt32FromByte(intb)
	offset := 4
	copy(b.parent[:], gobBlock[offset:offset+siacrypto.TruncatedHashSize])
	offset += siacrypto.TruncatedHashSize

	// decode each heartbeat
	var nextOffset uint32
	var heartbeatOffset uint32
	var i int
	for i = 0; i < quorum.QuorumSize-1; i++ {
		copy(intb[:], gobBlock[offset:])
		heartbeatOffset = siaencoding.UInt32FromByte(intb)
		offset += 4
		if heartbeatOffset == ^uint32(0) {
			b.heartbeats[i] = nil
			continue
		}
		copy(intb[:], gobBlock[offset:])
		nextOffset = siaencoding.UInt32FromByte(intb)

		j := 1
		for nextOffset == ^uint32(0) && j+i < quorum.QuorumSize {
			copy(intb[:], gobBlock[offset+4*j:])
			nextOffset = siaencoding.UInt32FromByte(intb)
			j++
		}

		if nextOffset == ^uint32(0) {
			// current heartbeat is last heartbeat
			break
		}

		if heartbeatOffset > nextOffset || int(nextOffset)+MinHeartbeatSize > len(gobBlock) {
			err = fmt.Errorf("Received invalid block")
			return
		}

		b.heartbeats[i].GobDecode(gobBlock[heartbeatOffset:nextOffset])
	}
	b.heartbeats[i].GobDecode(gobBlock[heartbeatOffset:])

	return
}

func (p *Participant) SaveBlock(b *block) (err error) {
	var file *os.File
	if p.activeHistoryStep == SnapshotLen {
		p.activeHistoryStep = 0
		os.Remove(p.recentHistory)
		p.recentHistory = p.activeHistory
		p.activeHistory = p.quorum.GetWalletPrefix()
		p.activeHistory += fmt.Sprintf(".blockHistory.%v", b.height)
		file, err = os.Create(p.activeHistory)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		p.activeHistoryStep += 1
		file, err = os.OpenFile(p.activeHistory, os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	blockHistoryHeaderBytes := make([]byte, BlockHistoryHeaderSize)
	n, err := file.Read(blockHistoryHeaderBytes)
	if err != nil || n != BlockHistoryHeaderSize {
		panic(err)
	}

	var bhh blockHistoryHeader
	err = bhh.GobDecode(blockHistoryHeaderBytes)
	if err != nil {
		panic(err)
	}

	gobBlock, err := b.GobEncode()
	if err != nil {
		return
	}

	if bhh.latestBlock == 0 {
		bhh.blockOffsets[0] = uint32(BlockHistoryHeaderSize)
		bhh.blockOffsets[1] = uint32(BlockHistoryHeaderSize)
	}

	if bhh.latestBlock != SnapshotLen-1 {
		bhh.blockOffsets[bhh.latestBlock+1] += uint32(len(gobBlock))
	}

	_, err = file.Seek(int64(bhh.blockOffsets[bhh.latestBlock]), 0)
	if err != nil {
		panic(err)
	}

	n, err = file.Write(gobBlock)
	if err != nil || n != len(gobBlock) {

	}

	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}

	encodeBHH, err := bhh.GobEncode()
	if err != nil {
		panic(err)
	}

	n, err = file.Write(encodeBHH)
	if err != nil || n != len(encodeBHH) {
		panic(err)
	}

	return
}
