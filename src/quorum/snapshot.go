package quorum

import (
	"fmt"
	"os"
	"siaencoding"
)

const (
	SnapHeaderSize = 8
)

type walletLookup struct {
	id     WalletID
	offset uint32
}

type snapshotHeader struct {
	walletLookupOffset uint32
	wallets            uint32
}

func (s *snapshotHeader) GobEncode() (b []byte, err error) {
	if s == nil {
		err = fmt.Errorf("cannot encode nil snapshotHeader")
		return
	}

	b = make([]byte, SnapHeaderSize)
	intb := siaencoding.EncUint32(s.walletLookupOffset)
	copy(b, intb[:])
	intb = siaencoding.EncUint32(s.wallets)
	copy(b[4:], intb[:])
	return
}

func (s *snapshotHeader) GobDecode(b []byte) (err error) {
	if s == nil {
		err = fmt.Errorf("cannode decode into nil snapshotHeader")
		return
	}
	if len(b) != SnapHeaderSize {
		err = fmt.Errorf("received invalid snap header")
		return
	}

	s.walletLookupOffset = siaencoding.DecUint32(b[:4])
	s.wallets = siaencoding.DecUint32(b[4:])
	return
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
	size, err := file.Write(q.loadWallet(w.id).bytes()[:])
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

// SaveSnap takes all of the variables in quorum and stores them to disk, such
// that anyone downloading the quorum and the blocks that follow the quorum can
// maintain a consistent state. First the quorum is saved, then a list of
// wallets, and then the actual wallets.
func (q *Quorum) SaveSnap() {
	// q.currentSnap is used to determine whether the current snapshot is snap0
	// or snap1. Each time SaveSnap is called, the snapshot is cycled to the
	// other one, meaning that there are 2 snaps at all times.
	q.currentSnapshot = !q.currentSnapshot
	snapname := q.walletPrefix
	if q.currentSnapshot {
		snapname += "snapshot0"
	} else {
		snapname += "snapshot1"
	}

	// create a new snapshot of the filename, obliterating the old snapshot of the
	// same filename
	file, err := os.Create(snapname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gobQuorum, err := q.GobEncode()
	if err != nil {
		panic(err)
	}

	// create the snapshot header and write it to disk
	header := snapshotHeader{
		walletLookupOffset: uint32(len(gobQuorum) + SnapHeaderSize),
		wallets:            q.wallets,
	}
	headerBytes, err := header.GobEncode()
	if err != nil {
		panic(err)
	}
	size, err := file.Write(headerBytes)
	if err != nil {
		panic(err)
	}
	offset := uint32(size)

	// write the encoded quorum to disk
	size, err = file.Write(gobQuorum)
	if err != nil {
		panic(err)
	}
	offset += uint32(size)

	// save an array indicating each wallet and its offset in the file. The
	// offsets are left blank for the time being and will be filled out when the
	// wallets are saved to disk.
	walletSliceBytes := make([]byte, q.wallets*12)
	size, err = file.Write(walletSliceBytes)
	offset += uint32(size)

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
	println(len(quorumBytes))
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
