package quorum

import (
	"fmt"
	"os"
)

type walletLookup struct {
	id     WalletID
	offset int
}

type WalletFetch struct {
	snap bool
	ids  []WalletID
}

// saveWalletTree goes through in sorted order and saves the wallets to disk.
// upon saving the wallets, an element is appended to the wallet index, which
// contains a list of all wallets and their offset in the snapshot. This only
// exists to enable linear lookup of individual wallets within the snapshot.
func (q *Quorum) saveWalletTree(w *walletNode, file *os.File, index *int, offset *int, walletSlice []walletLookup) {
	if w == nil {
		return
	}

	q.saveWalletTree(w.children[0], file, index, offset, walletSlice)
	q.saveWalletTree(w.children[1], file, index, offset, walletSlice)

	size, err := file.Write(q.loadWallet(w.id).bytes()[:])
	if err != nil {
		panic(err)
	}

	walletSlice[*offset] = walletLookup{
		id:     w.id,
		offset: *offset,
	}
	*index += 1
	*offset += size
	return
}

// Things saved in a Snap:
//
// 1. The quorum struct
// 2. A list of the wallets and their offsets
// 3 A list of the wallets and their scripts
func (q *Quorum) SaveSnap() {
	// open the file in which the snapshot is stored
	q.currentSnap = !q.currentSnap
	var snap int
	snapname := q.walletPrefix
	if q.currentSnap {
		snapname += ".snap0"
		snap = 0
	} else {
		snapname += ".snap1"
		snap = 1
	}
	file, err := os.Create(snapname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var offset int
	var index int
	walletSlice := make([]walletLookup, q.numNodes)

	// save quorum to disk
	gobQuorum, err := q.GobEncode()
	if err != nil {
		panic(err)
	}
	size, err := file.Write(gobQuorum)
	if err != nil {
		panic(err)
	}
	offset += size

	q.snapWalletSliceOffset[snap] = offset

	walletSliceBytes := make([]byte, q.numNodes*12)
	size, err = file.Write(walletSliceBytes) // create a placeholder in the file
	_, err = file.Seek(int64(len(walletSlice)), 1)
	if err != nil {
		panic(err)
	}
	offset += len(walletSlice)

	// get every wallet, and get its bytes
	q.saveWalletTree(q.walletRoot, file, &index, &offset, walletSlice)

	for i := range walletSlice {
		tmp64 := uint64(walletSlice[i].id)
		for j := 0; j < 8; j++ {
			walletSliceBytes[i*12+j] = byte(tmp64)
			tmp64 = tmp64 >> 8
		}

		tmp := walletSlice[i].offset
		for j := 8; j < 12; j++ {
			walletSliceBytes[i*12+j] = byte(tmp)
			tmp = tmp >> 8
		}
	}

	_, err = file.Seek(int64(q.snapWalletSliceOffset[snap]), 0)
	if err != nil {
		panic(err)
	}
	size, err = file.Write(walletSliceBytes)

	q.snapSize[snap] = offset
}

// loads and transfers the quorum componenet from the most recent snapshot
func (self *Quorum) FetchSnapQuorum(_ bool, q *Quorum) (err error) {
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
