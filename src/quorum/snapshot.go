package quorum

import(
	"fmt"
)

type walletLookup struct {
	id WalletID
	offset int
}

// saveWalletTree goes through in sorted order and saves the wallets to disk.
// upon saving the wallets, an element is appended to the wallet index, which
// contains a list of all wallets and their offset in the snapshot. This only
// exists to enable linear lookup of individual wallets within the snapshot.
func (q *Quorum) saveWalletTree(w *walletNode, file *os.File, index *int, offset *int, walletSlice []walletLookup) {
	if w == nil {
		return
	}

	q.saveWalletTree(w.children[0], file)
	q.saveWalletTree(w.children[1], file)

	size, err := file.Write(q.loadWallet(w.id).bytes()[:])
	if err != nil {
		panic(err)
	}

	walletSlice[offset] = walletLookup {
		id: w.id
		offset: offset
	}
	index += 1
	offset += size
	return
}

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
	size, err := file.Write(walletSlice) // create a placeholder in the file
	file.Seek(len(walletSlice), 1)
	offset += len(walletSlice)

	// get every wallet, and get its bytes
	q.saveWalletTree(q.walletRoot, file, &index, &offset, &walletSlice)

	file.Seek(walletSliceOffset, 0)
	size, err := file.Write(walletSlice)

	q.snapSize[snap] = offset
}

// loads and transfers the quorum componenet from the most recent snapshot
func (q *Quorum) FetchSnapQuorum(_ bool, q *Quorum) (err error) {
	snapname := q.walletPrefix
	var snap int
	if q.currentSnap {
		snapname += ".snap0"
		snap = 0
	} else {
		snapmane += ".snap1"
		snap = 1
	}

	file, err := os.Open(snapname)
	if err != nil {
		return
	}
	defer file.Close()

	quorumBytes := make([]byte,  q.snapWalletSliceOffset[snap])
	n, err := file.Read(quorumBytes)
	if err != nil || n != len(quorumBytes) {
		err = fmt.Errorf("error reading snapshot into memory")
		return
	}

	err = q.GobDecode(quorumBytes)
	return
}
}
