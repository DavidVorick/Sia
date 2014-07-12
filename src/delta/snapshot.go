package delta

import (
	"fmt"
	"os"
	"siaencoding"
)

const (
	SnapshotHeaderSize        = 8
	snapshotOffsetTableLength = 24
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
	quorumMetaDataOffset uint32
	quorumMetaDataLength uint32

	walletLookupTableOffset uint32
	walletLookupTableLength uint32

	eventLookupTableOffset uint32
	eventLookupTableLength uint32
}

type walletOffset struct {
	id     quorum.WalletID
	offset uint32
	length uint32
}

// saveWalletTree goes through in sorted order and saves the wallets to disk.
// upon saving the wallets, an element is appended to the wallet index, which
// contains a list of all wallets and their offset in the snapshot. This only
// exists to enable linear lookup of individual wallets within the snapshot.
func (q *Quorum) saveWalletTree(w *walletNode, file *os.File, index *int, offset *uint32, walletSlice []walletLookup) {
	if w == nil {
		return
	}

	// save all wallets that are less than the current wallet
	q.saveWalletTree(w.children[0], file, index, offset, walletSlice)

	// save the current wallet
	wBytes, err := q.LoadWallet(w.id).GobEncode()
	if err != nil {
		panic(err)
	}
	size, err := file.Write(wBytes)
	if err != nil {
		panic(err)
	}
	walletSlice[*index] = walletLookup{
		id:     w.id,
		offset: *offset,
	}
	*index += 1
	*offset += uint32(size)

	// save all wallets greater than the current wallet
	q.saveWalletTree(w.children[1], file, index, offset, walletSlice)
	return
}

// SaveSnapshot takes all of the variables listed at the top of the file,
// encodes them, and writes to disk.
func (e *Engine) SaveSnapshot() (err error) {
	// Determine the filename for the snapshot
	snapshotFilename := fmt.Sprintf("%s.snapshot.%v", e.filePrefix, e.activeHistoryHead)

	file, err := os.Create(snapshotFilename)
	if err != nil {
		return
	}
	defer file.Close()

	// List of offsets that prefix the snapshot file
	var offsetTable snapshotOffsetTable

	// put encodedQuorumMetaData in it's own scope so it can be cleared before
	// the function returns
	{
		// encode the quorum and record the length
		encodedQuorumMetaData, err := e.quorum.MarshalMetaData()
		if err != nil {
			return
		}
		offsetTable.quorumMetaDataSize = len(encodedQuorum)
		offsetTable.quorumMetaDataOffset = snapshotOffsetTableLength

		// Write the encoded quorum to the snapshot file.
		_, err = file.Seek(int64(offsetTable.quorumMetaDataOffset), 0)
		if err != nil {
			return
		}
		_, err = file.Write(encodedQuorumMetaData)
		if err != nil {
			return
		}
	}

	// create the wallet lookup table and save the wallets
	walletDigest := e.quorum.WalletDigest()

	// save every wallet to disk, recording the offset and id in the wallet lookup
	// array at the beginning of the file
	var index int
	walletSlice := make([]walletLookup, q.wallets)
	q.saveWalletTree(q.walletRoot, file, &index, &offset, walletSlice)

	// fill out walletSliceBytes with the wallet lookup table
	for i := range walletSlice {
		intb := siaencoding.EncUint64(uint64(walletSlice[i].id))
		copy(walletSliceBytes[i*12:], intb[:])
		int32b := siaencoding.EncUint32(walletSlice[i].offset)
		copy(walletSliceBytes[i*12+8:], int32b[:])
	}

	// seek to the offset where the wallet lookup table is kept and save the table
	_, err = file.Seek(int64(header.walletLookupOffset), 0)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(walletSliceBytes)
	if err != nil {
		panic(err)
	}
}

// loads and transfers the quorum componenet from the most recent snapshot
func (self *Quorum) RecentSnapshot() (q *Quorum, err error) {
	snapname := self.walletPrefix
	if self.currentSnapshot {
		snapname += "snapshot0"
	} else {
		snapname += "snapshot1"
	}

	file, err := os.Open(snapname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var header snapshotHeader
	headerBytes := make([]byte, SnapHeaderSize)
	n, err := file.Read(headerBytes)
	if n != len(headerBytes) || err != nil {
		panic(err)
	}
	err = header.GobDecode(headerBytes)
	if err != nil {
		panic(err)
	}

	q = new(Quorum)
	quorumBytes := make([]byte, header.walletLookupOffset-SnapHeaderSize)
	n, err = file.Read(quorumBytes)
	if err != nil || n != len(quorumBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	err = q.GobDecode(quorumBytes)
	if err != nil {
		panic(err)
	}
	return
}

// returns the list of all wallets in a given snapshot
func (q *Quorum) SnapshotWalletList(snap bool) (ids []WalletID) {
	snapname := q.walletPrefix
	if snap {
		snapname += "snapshot0"
	} else {
		snapname += "snapshot1"
	}

	file, err := os.Open(snapname)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	var header snapshotHeader
	headerBytes := make([]byte, SnapHeaderSize)
	n, err := file.Read(headerBytes)
	if n != len(headerBytes) || err != nil {
		panic(err)
	}
	err = header.GobDecode(headerBytes)
	if err != nil {
		panic(err)
	}

	_, err = file.Seek(int64(header.walletLookupOffset), 0)
	if err != nil {
		panic(err)
	}

	lookupBytes := make([]byte, header.wallets*12)
	n, err = file.Read(lookupBytes)
	if err != nil || n != len(lookupBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	ids = make([]WalletID, header.wallets)
	for i := uint32(0); i < header.wallets; i++ {
		ids[i] = WalletID(siaencoding.DecUint64(lookupBytes[i*12 : i*12+8]))
	}

	return
}

func (q *Quorum) SnapshotWallets(snap bool, ids []WalletID) (encodedWallets [][]byte) {
	snapname := q.walletPrefix
	if snap {
		snapname += "snapshot0"
	} else {
		snapname += "snapshot1"
	}

	file, err := os.Open(snapname)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	// This is the third time this set of code appears in the file, should
	// probably be a getHeader(file) function.
	var header snapshotHeader
	headerBytes := make([]byte, SnapHeaderSize)
	n, err := file.Read(headerBytes)
	if n != len(headerBytes) || err != nil {
		panic(err)
	}
	err = header.GobDecode(headerBytes)
	if err != nil {
		panic(err)
	}

	_, err = file.Seek(int64(header.walletLookupOffset), 0)
	if err != nil {
		panic(err)
	}

	lookupBytes := make([]byte, header.wallets*12)
	n, err = file.Read(lookupBytes)
	if err != nil || n != len(lookupBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	lookup := make([]walletLookup, header.wallets)
	for i := uint32(0); i < header.wallets; i++ {
		lookup[i].id = WalletID(siaencoding.DecUint64(lookupBytes[i*12 : i*12+8]))
		lookup[i].offset = siaencoding.DecUint32(lookupBytes[i*12+8 : i*12+12])
	}

	// find each wallet and add it to the list of encoded wallets
	encodedWallets = make([][]byte, len(ids))
	for i, id := range ids {
		// wallet lookup is sorted; can do a binary search
		low := 0
		high := len(lookup) - 1
		mid := 0
		for high >= low {
			mid = (low + high) / 2
			if lookup[mid].id == id {
				break
			}
			if id > lookup[mid].id {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}

		if lookup[mid].id != id {
			encodedWallets[i] = nil
			continue
		}

		// fetch the wallet from disk
		_, err = file.Seek(int64(lookup[mid].offset), 0)
		if err != nil {
			panic(err)
		}

		fileInfo, err := file.Stat()
		if err != nil {
			panic(err)
		}
		fileSize := fileInfo.Size()

		var walletSize int64
		if mid == len(lookup)-1 {
			walletSize = fileSize - int64(lookup[mid].offset)
		} else {
			walletSize = int64(lookup[mid+1].offset - lookup[mid].offset)
		}
		encodedWallet := make([]byte, walletSize)
		_, err = file.Read(encodedWallet)
		if err != nil {
			panic(err)
		}
		encodedWallets[i] = encodedWallet
	}

	return
}
