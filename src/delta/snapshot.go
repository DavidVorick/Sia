package delta

import (
	"fmt"
	"os"
	"siaencoding"
	"state"
)

const (
	SnapshotHeaderSize        = 8
	snapshotOffsetTableLength = 24
	walletOffsetLength        = 16
)

// A snapshot is an on-disk representation of a quorum. A snapshot is not meant
// to be sent over a wire all at once, but instead pieces of a snapshot can be
// requested one at a time. This is because a full snapshot could be quite
// large and costly to have in memory all at once.
//
// The layout of a snapshot is as follows:
// 1. A table at the front of the file describing the location of each structure, prefixed by its size
//		1a. Offset of quorum meta data + size of quorum meta data
//		1b. Offset of wallet lookup table + size of wallet lookup table
//		1c. Offset of event lookup table + size of event lookup table
// 2. Quorum meta data
// 3. Wallet lookup table
// 4. Wallets with their scripts
// 5. Event lookup table
// 6. Events

type snapshotOffsetTable struct {
	stateMetadataOffset uint32
	stateMetadataLength uint32

	walletLookupTableOffset uint32
	walletLookupTableLength uint32

	eventLookupTableOffset uint32
	eventLookupTableLength uint32
}

func (s *snapshotOffsetTable) encode() (b []byte, err error) {
	b = make([]byte, snapshotOffsetTableLength)
	var offset int
	qmdo := siaencoding.EncUint32(s.stateMetadataOffset)
	copy(b[offset:], qmdo)
	offset += 4
	qmdl := siaencoding.EncUint32(s.stateMetadataLength)
	copy(b[offset:], qmdl)
	offset += 4
	wlto := siaencoding.EncUint32(s.walletLookupTableOffset)
	copy(b[offset:], wlto)
	offset += 4
	wltl := siaencoding.EncUint32(s.walletLookupTableLength)
	copy(b[offset:], wltl)
	offset += 4
	elto := siaencoding.EncUint32(s.eventLookupTableOffset)
	copy(b[offset:], elto)
	offset += 4
	eltl := siaencoding.EncUint32(s.eventLookupTableLength)
	copy(b[offset:], eltl)
	return
}

func (s *snapshotOffsetTable) decode(b []byte) (err error) {
	if len(b) <= snapshotOffsetTableLength {
		err = fmt.Errorf("snapshotOffsetTable decode: input is too small to contain a snapshotOffsetTable.")
		return
	}

	var offset int
	s.stateMetadataOffset = siaencoding.DecUint32(b[offset:])
	offset += 4
	s.stateMetadataLength = siaencoding.DecUint32(b[offset:])
	offset += 4
	s.walletLookupTableOffset = siaencoding.DecUint32(b[offset:])
	offset += 4
	s.walletLookupTableLength = siaencoding.DecUint32(b[offset:])
	offset += 4
	s.eventLookupTableOffset = siaencoding.DecUint32(b[offset:])
	offset += 4
	s.eventLookupTableLength = siaencoding.DecUint32(b[offset:])
	return
}

type walletOffset struct {
	id     state.WalletID
	offset uint32
	length uint32
}

func (wo *walletOffset) encode() (b []byte, err error) {
	b = make([]byte, walletOffsetLength)
	var offset int
	encID := siaencoding.EncUint64(uint64(wo.id))
	copy(b[offset:], encID)
	offset += 8
	encOffset := siaencoding.EncUint32(wo.offset)
	copy(b[offset:], encOffset)
	offset += 4
	encLength := siaencoding.EncUint32(wo.length)
	copy(b[offset:], encLength)
	return
}

func (wo *walletOffset) decode(b []byte) (err error) {
	if len(b) <= walletOffsetLength {
		err = fmt.Errorf("walletOffset decode: input is too small to contain a walletOffset")
		return
	}

	var offset int
	wo.id = state.WalletID(siaencoding.DecUint64(b[offset:]))
	offset += 8
	wo.offset = siaencoding.DecUint32(b[offset:])
	offset += 4
	wo.length = siaencoding.DecUint32(b[offset:])
	return
}

func (e *Engine) snapshotFilename(height uint32) (snapshotFilename string) {
	snapshotFilename = fmt.Sprintf("%s.snapshot.%v", e.filePrefix, height)
	return
}

func (e *Engine) openSnapshot(snapshotHead uint32) (file *os.File, snapshotTable snapshotOffsetTable, err error) {
	// Make sure that the requested snapshot is on disk.
	if snapshotHead == ^uint32(0) || (snapshotHead != e.activeHistoryHead && snapshotHead != e.recentHistoryHead) {
		err = fmt.Errorf("Snapshot not found.")
		return
	}

	// Open the file associated with the requested snapshot.
	snapshotFilename := e.snapshotFilename(snapshotHead)
	file, err = os.Open(snapshotFilename)
	if err != nil {
		return
	}

	// Read and decode the snapshot offset table.
	encodedSnapshotTable := make([]byte, snapshotOffsetTableLength)
	_, err = file.Read(encodedSnapshotTable)
	if err != nil {
		return
	}
	var decodedSnapshotTable snapshotOffsetTable
	err = decodedSnapshotTable.decode(encodedSnapshotTable)
	return
}

// SaveSnapshot takes all of the variables listed at the top of the file,
// encodes them, and writes to disk.
func (e *Engine) SaveSnapshot() (err error) {
	// Determine the filename for the snapshot
	snapshotFilename := e.snapshotFilename(e.activeHistoryHead)
	file, err := os.Create(snapshotFilename)
	if err != nil {
		return
	}
	defer file.Close()

	// List of offsets that prefix the snapshot file
	var offsetTable snapshotOffsetTable
	currentOffset := snapshotOffsetTableLength

	// put encodedQuorumMetadata in it's own scope so it can be cleared before
	// the function returns
	{
		// encode the quorum and record the length
		var encodedQuorumMetadata []byte
		encodedQuorumMetadata, err = siaencoding.Marshal(e.state.Metadata)
		if err != nil {
			return
		}
		offsetTable.stateMetadataLength = uint32(len(encodedQuorumMetadata))
		offsetTable.stateMetadataOffset = uint32(currentOffset)

		// Write the encoded quorum to the snapshot file.
		_, err = file.Seek(int64(offsetTable.stateMetadataOffset), 0)
		if err != nil {
			return
		}
		_, err = file.Write(encodedQuorumMetadata)
		if err != nil {
			return
		}
		currentOffset += len(encodedQuorumMetadata)
	}

	// Create the wallet lookup table and save the wallets. This is again in its
	// own scope so that the data can be garbage collected before the function
	// returns.
	{
		// Retreive a list of all the wallets stored in the quorum and allocate the wallet lookup table
		walletList := e.state.WalletList()
		offsetTable.walletLookupTableOffset = uint32(currentOffset)
		offsetTable.walletLookupTableLength = uint32(len(walletList) * walletOffsetLength)
		walletLookupTable := make([]walletOffset, len(walletList))
		currentOffset += len(walletList) * walletOffsetLength

		// Write wallets, update lookup table.
		for i := range walletList {
			var encodedWallet []byte
			var wallet state.Wallet
			wallet, err = e.state.LoadWallet(walletList[i])
			if err != nil {
				return
			}
			encodedWallet, err = siaencoding.Marshal(wallet)
			if err != nil {
				return
			}
			walletLookupTable[i].length = uint32(len(encodedWallet))
			walletLookupTable[i].offset = uint32(currentOffset)
			_, err = file.Write(encodedWallet)
			if err != nil {
				return
			}
			currentOffset += len(encodedWallet)
		}

		// Encode lookup table.
		encodedWalletLookupTable := make([]byte, len(walletLookupTable)*walletOffsetLength)
		for i := range walletLookupTable {
			var encodedLookup []byte
			encodedLookup, err = walletLookupTable[i].encode()
			if err != nil {
				return
			}
			copy(encodedWalletLookupTable[i*walletOffsetLength:], encodedLookup)
		}

		// Write lookup table.
		_, err = file.Seek(int64(offsetTable.walletLookupTableOffset), 0)
		if err != nil {
			return
		}
		_, err = file.Write(encodedWalletLookupTable)
		if err != nil {
			return
		}
	}

	// event list stuff here

	// Encode and write 'offsetTable'
	encodedOffset, err := offsetTable.encode()
	if err != nil {
		return
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}
	_, err = file.Write(encodedOffset)
	return
}

func (e *Engine) LoadSnapshotMetadata(snapshotHead uint32) (snapshot state.StateMetadata, err error) {
	// Open the file holding the desired snapshot. This function also provides
	// the snapshot table.
	file, snapshotTable, err := e.openSnapshot(snapshotHead)
	if err != nil {
		return
	}
	defer file.Close()

	// Determine length and offset of metadata, then load and decode the metadata.
	encodedSnapshotMetadata := make([]byte, snapshotTable.stateMetadataLength)
	_, err = file.Seek(int64(snapshotTable.stateMetadataOffset), 0)
	if err != nil {
		return
	}
	_, err = file.Read(encodedSnapshotMetadata)
	if err != nil {
		return
	}
	err = siaencoding.Unmarshal(encodedSnapshotMetadata, &snapshot)
	return
}

func (e *Engine) LoadWalletList(snapshotHead uint32) (walletList []state.WalletID, err error) {
	// Open the snapshot.
	file, snapshotTable, err := e.openSnapshot(snapshotHead)
	if err != nil {
		return
	}
	defer file.Close()

	// Determine the length and offset of the wallet table, then load it.
	encodedWalletOffsetTable := make([]byte, snapshotTable.walletLookupTableLength)
	_, err = file.Seek(int64(snapshotTable.walletLookupTableOffset), 0)
	if err != nil {
		return
	}
	_, err = file.Read(encodedWalletOffsetTable)
	if err != nil {
		return
	}

	// Decode the wallet lookup table into walletList.
	for i := uint32(0); i < snapshotTable.walletLookupTableLength; i += walletOffsetLength {
		var wo walletOffset
		err = wo.decode(encodedWalletOffsetTable[i*walletOffsetLength : i*walletOffsetLength+walletOffsetLength])
		if err != nil {
			return
		}
		walletList = append(walletList, wo.id)
	}

	return
}
