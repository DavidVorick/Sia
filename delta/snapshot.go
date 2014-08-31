package delta

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

const (
	// SnapshotHeaderSize is the size, in bytes, of the snapshot header.
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
// 1. A table at the front of the file describing the location of each
// structure, prefixed by its size.
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
	buf := new(bytes.Buffer)
	buf.Write(siaencoding.EncUint32(s.stateMetadataOffset))
	buf.Write(siaencoding.EncUint32(s.stateMetadataLength))
	buf.Write(siaencoding.EncUint32(s.walletLookupTableOffset))
	buf.Write(siaencoding.EncUint32(s.walletLookupTableLength))
	buf.Write(siaencoding.EncUint32(s.eventLookupTableOffset))
	buf.Write(siaencoding.EncUint32(s.eventLookupTableLength))
	b = buf.Bytes()
	return
}

func (s *snapshotOffsetTable) decode(b []byte) (err error) {
	if len(b) < snapshotOffsetTableLength {
		err = errors.New("input is too small to contain a snapshotOffsetTable")
		return
	}

	buf := bytes.NewBuffer(b)
	s.stateMetadataOffset = siaencoding.DecUint32(buf.Next(4))
	s.stateMetadataLength = siaencoding.DecUint32(buf.Next(4))
	s.walletLookupTableOffset = siaencoding.DecUint32(buf.Next(4))
	s.walletLookupTableLength = siaencoding.DecUint32(buf.Next(4))
	s.eventLookupTableOffset = siaencoding.DecUint32(buf.Next(4))
	s.eventLookupTableLength = siaencoding.DecUint32(buf.Next(4))
	return
}

type walletOffset struct {
	id     state.WalletID
	offset uint32
	length uint32
}

func (wo *walletOffset) encode() (b []byte, err error) {
	buf := new(bytes.Buffer)
	buf.Write(siaencoding.EncUint64(uint64(wo.id)))
	buf.Write(siaencoding.EncUint32(wo.offset))
	buf.Write(siaencoding.EncUint32(wo.length))
	b = buf.Bytes()
	return
}

func (wo *walletOffset) decode(b []byte) (err error) {
	if len(b) < walletOffsetLength {
		err = errors.New("input is too small to contain a walletOffset")
		return
	}

	buf := bytes.NewBuffer(b)
	wo.id = state.WalletID(siaencoding.DecUint64(buf.Next(8)))
	wo.offset = siaencoding.DecUint32(buf.Next(4))
	wo.length = siaencoding.DecUint32(buf.Next(4))
	return
}

func (e *Engine) snapshotFilename(height uint32) (snapshotFilename string) {
	snapshotFilename = fmt.Sprintf("%ssnapshot.%v", e.filePrefix, height)
	return
}

// saveSnapshot takes all of the variables listed at the top of the file,
// encodes them, and writes to disk.
func (e *Engine) saveSnapshot() (err error) {
	fmt.Println("Saving Snapshot")
	fmt.Println(e.siblingIndex)
	fmt.Println(e.state.Metadata.Height)
	fmt.Println(e.state.Metadata.RecentSnapshot)
	fmt.Println(e.activeHistoryLength)
	// Determine the filename for the snapshot
	snapshotFilename := e.snapshotFilename(e.state.Metadata.Height)
	file, err := os.Create(snapshotFilename)
	if err != nil {
		return
	}
	defer file.Close()

	// Set the new value for RecentSnapshot. This needs to be set before
	// the snapshot is saved because the value of the recent snapshot needs
	// to get included in the metadata during the save.
	e.state.Metadata.RecentSnapshot = e.state.Metadata.Height

	// List of offsets that prefix the snapshot file
	var offsetTable snapshotOffsetTable
	currentOffset := snapshotOffsetTableLength

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
		_, err = file.WriteAt(encodedQuorumMetadata, int64(offsetTable.stateMetadataOffset))
		if err != nil {
			return
		}
		currentOffset += len(encodedQuorumMetadata)
	}

	// Create wallet lookup table and save the wallets.
	{
		// Retreive a list of all the wallets stored in the quorum and allocate
		// the wallet lookup table
		walletList := e.state.WalletList()
		offsetTable.walletLookupTableOffset = uint32(currentOffset)
		offsetTable.walletLookupTableLength = uint32(len(walletList) * walletOffsetLength)
		walletLookupTable := make([]walletOffset, len(walletList))
		currentOffset += len(walletList) * walletOffsetLength

		// Seek the file to the current offset, since the offset was changed
		// without doing a file write.
		_, err = file.Seek(int64(currentOffset), 0)
		if err != nil {
			return
		}

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
			walletLookupTable[i].id = wallet.ID
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
		_, err = file.WriteAt(encodedWalletLookupTable, int64(offsetTable.walletLookupTableOffset))
		if err != nil {
			return
		}
	}

	// Encode and write 'offsetTable'
	encodedOffset, err := offsetTable.encode()
	if err != nil {
		return
	}
	_, err = file.WriteAt(encodedOffset, 0)
	if err != nil {
		return
	}

	// Delete the oldest snapshot.
	oldSnapshotFilename := e.snapshotFilename(e.state.Metadata.RecentSnapshot - 2*SnapshotLength)
	err = siafiles.Remove(oldSnapshotFilename)
	if err != nil {
		return
	}

	return
}

func (e *Engine) openSnapshot(snapshotHead uint32) (file *os.File, snapshotTable snapshotOffsetTable, err error) {
	// Make sure that the requested snapshot is on disk.
	if snapshotHead == ^uint32(0) || (snapshotHead != e.state.Metadata.RecentSnapshot && snapshotHead != e.recentHistoryHead) {
		err = fmt.Errorf("snapshot %v not found", snapshotHead)
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
	err = snapshotTable.decode(encodedSnapshotTable)
	return
}

// LoadSnapshotMetadata returns the Metadata object corresponding to a given
// snapshot head.
func (e *Engine) LoadSnapshotMetadata(snapshotHead uint32) (snapshot state.Metadata, err error) {
	// Open the file holding the desired snapshot. This function also provides
	// the snapshot table.
	file, snapshotTable, err := e.openSnapshot(snapshotHead)
	if err != nil {
		return
	}
	defer file.Close()

	// Determine length and offset of metadata, then load and decode the
	// metadata.
	encodedSnapshotMetadata := make([]byte, snapshotTable.stateMetadataLength)
	_, err = file.ReadAt(encodedSnapshotMetadata, int64(snapshotTable.stateMetadataOffset))
	if err != nil {
		return
	}
	err = siaencoding.Unmarshal(encodedSnapshotMetadata, &snapshot)
	return
}

func (e *Engine) openWalletOffsetTable(snapshotHead uint32) (file *os.File, walletTable []byte, err error) {
	// Open the snapshot.
	file, snapshotTable, err := e.openSnapshot(snapshotHead)
	if err != nil {
		return
	}

	// Determine the length and offset of the wallet table, then load it.
	walletTable = make([]byte, snapshotTable.walletLookupTableLength)
	_, err = file.ReadAt(walletTable, int64(snapshotTable.walletLookupTableOffset))
	return
}

// LoadSnapshotWalletList returns the list of WalletIDs corresponding to a
// given snapshot head.
func (e *Engine) LoadSnapshotWalletList(snapshotHead uint32) (walletList []state.WalletID, err error) {
	// Get the wallet table and snapshot file.
	file, walletTable, err := e.openWalletOffsetTable(snapshotHead)
	if err != nil {
		return
	}
	defer file.Close()

	// Decode the wallet lookup table into walletList.
	for i := 0; i < len(walletTable); i += walletOffsetLength {
		var wo walletOffset
		err = wo.decode(walletTable[i:])
		if err != nil {
			return
		}
		walletList = append(walletList, wo.id)
	}

	return
}

// LoadSnapshotWallet returns the Wallet object corresponding to a given
// WalletID at a given snapshot head.
func (e *Engine) LoadSnapshotWallet(snapshotHead uint32, walletID state.WalletID) (wallet state.Wallet, err error) {
	// Open the wallet table and snapshot file.
	file, walletTable, err := e.openWalletOffsetTable(snapshotHead)
	if err != nil {
		return
	}
	defer file.Close()

	// Determine the offset of the wallet in question, via binary search.
	max := len(walletTable)/walletOffsetLength - 1
	min := 0
	for max >= min {
		mid := (max + min) / 2

		// Load the wallet associated with the midpoint.
		var midWalletOffset walletOffset
		err = midWalletOffset.decode(walletTable[mid*walletOffsetLength:])
		if err != nil {
			return
		}
		midID := midWalletOffset.id

		// Determine which half of the remaining table the wallet resides within.
		if midID == walletID {
			// Fetch and decode the wallet.
			walletBytes := make([]byte, midWalletOffset.length)
			_, err = file.ReadAt(walletBytes, int64(midWalletOffset.offset))
			if err != nil {
				return
			}
			err = siaencoding.Unmarshal(walletBytes, &wallet)

			if wallet.KnownScripts == nil {
				println("nil map issue")
				wallet.KnownScripts = make(map[string]state.ScriptInputEvent)
			}
			return
		} else if midID < walletID {
			min = mid + 1
		} else {
			max = mid - 1
		}
	}

	err = errors.New("wallet is not stored within this snapshot")
	return
}
