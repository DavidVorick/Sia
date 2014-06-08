package participant

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siaencoding"
)

const (
	SnapshotLen            int = 100
	BlockHistoryHeaderSize int = 4 + SnapshotLen*4 + siacrypto.TruncatedHashSize*SnapshotLen
)

type block struct {
	hash       siacrypto.TruncatedHash
	height     int
	parent     siacrypto.TruncatedHash
	heartbeats [quorum.QuorumSize]*heartbeat
}

type blockHistoryHeader struct {
	latestBlock  int
	blockOffsets [SnapshotLen]int
}

func (bhh *blockHistoryHeader) GobEncode() (gobBHH []byte, err error) {
	gobBHH = make([]byte, BlockHistoryHeaderSize)
	encodedInt := siaencoding.IntToByte(bhh.latestBlock)
	copy(gobBHH, encodedInt[:])
	offset := 4

	for i := range bhh.blockOffsets {
		encodedInt = siaencoding.IntToByte(bhh.blockOffsets[i])
		copy(gobBHH[offset:], encodedInt[:])
		offset += 4
	}

	return
}

func (bhh *blockHistoryHeader) GobDecode(gobBHH []byte, err error) {
	if len(gobBHH) != BlockHistoryHeaderSize {
		err = fmt.Errorf("gobBHH has wrong size, cannot decode!")
	}

	var intb [4]byte
	copy(intb[:], gobBHH)
	bhh.latestBlock = siaencoding.IntFromByte(intb)
	offset := 4

	for i := range bhh.blockOffsets {
		copy(intb[:], gobBHH[offset:])
		bhh.blockOffsets[i] = siaencoding.IntFromByte(intb)
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
	intb := siaencoding.IntToByte(b.height)
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
	size += 4 * quorum.QuorumSize // an offset for each heartbeat

	// encode height, then parent, then heartbeats
	gobBlock = make([]byte, size)
	copy(gobBlock, intb[:])
	offset := len(intb)
	copy(gobBlock[offset:], b.parent[:])
	offset += siacrypto.TruncatedHashSize
	heartbeatOffset := offset + len(intb)*quorum.QuorumSize
	for i := range encodedHeartbeats {
		intb = siaencoding.IntToByte(len(encodedHeartbeats[i]))
		copy(gobBlock[offset:], intb[:])
		copy(gobBlock[heartbeatOffset:], encodedHeartbeats[i])
		offset += len(intb)
		heartbeatOffset += len(encodedHeartbeats[i])
	}

	return
}

func (b *block) GobDecode(gobBlock []byte) (err error) {
	if b == nil {
		err = fmt.Errorf("cannot decode into nil value")
		return
	}

	return
}

func (p *Participant) SaveBlock(b *block) (err error) {
	file, err := os.OpenFile(p.activeHistory, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	gobBlock, err := b.GobEncode()
	if err != nil {
		return
	}

	n, err := file.Write(gobBlock)
	if err != nil || n != len(gobBlock) {

	}
	return
}

func (p *Participant) LoadBlock() {
	//
	return
}
