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

type snapHeader struct {
	walletLookupOffset uint32
	wallets            uint32
}

func (s *snapHeader) GobEncode() (b []byte, err error) {
	if s == nil {
		err = fmt.Errorf("cannot encode nil snapHeader")
		return
	}

	b = make([]byte, SnapHeaderSize)
	intb := siaencoding.UInt32ToByte(s.walletLookupOffset)
	copy(b, intb[:])
	intb = siaencoding.UInt32ToByte(s.wallets)
	copy(b[4:], intb[:])
	return
}

func (s *snapHeader) GobDecode(b []byte, err error) {
	if s == nil {
		err = fmt.Errorf("cannode decode into nil snapHeader")
		return
	}
	if len(b) != SnapHeaderSize {
		err = fmt.Errorf("received invalid snap header")
		return
	}

	var intb [4]byte
	copy(intb[:], b)
	s.walletLookupOffset = siaencoding.UInt32FromByte(intb)
	copy(intb[:], b[4:])
	s.wallets = siaencoding.UInt32FromByte(intb)
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
	walletSlice[*offset] = walletLookup{
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
	var snap int
	q.currentSnap = !q.currentSnap
	snapname := q.walletPrefix
	if q.currentSnap {
		snapname += ".snap0"
		snap = 0
	} else {
		snapname += ".snap1"
		snap = 1
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
	header := snapHeader{
		walletLookupOffset: uint32(len(gobQuorum)),
		wallets:            q.numNodes,
	}
	headerBytes, err := header.GobEncode()
	if err != nil {
		panic(err)
	}
	size, err := file.Write(headerBytes)
	if err != nil {
		panic(err)
	}

	// write the encoded quorum to disk
	_, err = file.Write(gobQuorum)
	if err != nil {
		panic(err)
	}
	offset := uint32(size)

	// save an array indicating each wallet and its offset in the file. The
	// offsets are left blank for the time being and will be filled out when the
	// wallets are saved to disk.
	walletSliceBytes := make([]byte, q.numNodes*12)
	_, err = file.Write(walletSliceBytes)
	offset += uint32(size)

	// save every wallet to disk, recording the offset and id in the wallet lookup
	// array at the beginning of the file
	var index int
	walletSlice := make([]walletLookup, q.numNodes)
	q.saveWalletTree(q.walletRoot, file, &index, &offset, walletSlice)

	// fill out walletSliceBytes with the wallet lookup table
	for i := range walletSlice {
		intb := siaencoding.UInt64ToByte(uint64(walletSlice[i].id))
		copy(walletSliceBytes[i*12:], intb[:])
		int32b := siaencoding.UInt32ToByte(walletSlice[i].offset)
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
func (self *Quorum) FetchSnapQuorum() (q *Quorum, err error) {
	snapname := self.walletPrefix
	var snap int
	if self.currentSnap {
		snapname += ".snap0"
		snap = 0
	} else {
		snapname += ".snap1"
		snap = 1
	}

	file, err := os.Open(snapname)
	if err != nil {
		return
	}
	defer file.Close()

	quorumBytes := make([]byte, self.snapWalletSliceOffset[snap])
	n, err := file.Read(quorumBytes)
	if err != nil || n != len(quorumBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	q = new(Quorum)
	err = q.GobDecode(quorumBytes)
	return
}

// returns the list of all wallets in a given snapshot
func (q *Quorum) FetchSnapWalletList(snap bool, ids *[]WalletID) {
	snapname := q.walletPrefix
	var snapIndex int
	if snap {
		snapname += ".snap0"
		snapIndex = 0
	} else {
		snapname += ".snap1"
		snapIndex = 1
	}

	file, err := os.Open(snapname)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	_, err = file.Seek(int64(q.snapWalletSliceOffset[snapIndex]), 0)
	if err != nil {
		panic(err)
	}

	lookupBytes := make([]byte, q.snapWallets[snapIndex]*12)
	n, err := file.Read(lookupBytes)
	if err != nil || n != len(lookupBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	idsTmp := make([]WalletID, q.snapWallets[snapIndex])
	for i := 0; i < q.snapWallets[snapIndex]; i++ {
		for j := 7; j > 0; j-- {
			idsTmp[i] += WalletID(lookupBytes[i*12+j])
			idsTmp[i] = idsTmp[i] << 8
		}
		idsTmp[i] += WalletID(lookupBytes[i*12])
	}
	*ids = idsTmp
}

func (q *Quorum) FetchSnapWallets(wf WalletFetch, wallets *[][]byte) {
	snapname := q.walletPrefix
	var snapIndex int
	if wf.snap {
		snapname += ".snap0"
		snapIndex = 0
	} else {
		snapname += ".snap1"
		snapIndex = 1
	}

	file, err := os.Open(snapname)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	_, err = file.Seek(int64(q.snapWalletSliceOffset[snapIndex]), 0)
	if err != nil {
		panic(err)
	}

	lookupBytes := make([]byte, q.snapWallets[snapIndex]*12)
	n, err := file.Read(lookupBytes)
	if err != nil || n != len(lookupBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	lookups := make([]walletLookup, q.snapWallets[snapIndex])
	for i := 0; i < q.snapWallets[snapIndex]; i++ {
		for j := 7; j > 0; j-- {
			lookups[i].id += WalletID(lookupBytes[i*12+j])
			lookups[i].id = lookups[i].id << 8
		}
		lookups[i].id += WalletID(lookupBytes[i*12])

		for j := 11; j > 8; j-- {
			lookups[i].offset += int(lookupBytes[i*12+j])
			lookups[i].offset = lookups[i].offset << 8
		}
		lookups[i].offset += int(lookupBytes[i*12+8])
	}

	// binary serach and find all the wallets
}
