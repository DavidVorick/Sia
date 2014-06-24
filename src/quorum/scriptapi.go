package quorum

import (
	"errors"
	"os"
	"siacrypto"
)

// Cost structures...
// There's a computational cost associated with all of these actions, but there is also a storage cost.
// And there might also be other costs associated, such as network costs.
// I don't know the best way to handle oall of this

const (
	CreateWalletMaxCost = 8
	SendMaxCost         = 6
	AddSiblingMaxCost   = 50
)

// CreateWallet takes an id, a Balance, and an initial script and uses
// those to create a new wallet that gets stored in stable memory.
// If a wallet of that id already exists then the process aborts.
func (q *Quorum) CreateWallet(w *Wallet, id WalletID, balance Balance, initialScript []byte) (cost int, err error) {
	cost += 1
	if !w.Balance.Compare(balance) {
		err = errors.New("insufficient balance")
		return
	}

	// check if the new wallet already exists
	cost += 2
	wn := q.retrieve(id)
	if wn != nil {
		err = errors.New("wallet already exists")
		return
	}

	// create a wallet node to insert into the walletTree
	cost += 5
	wn = new(walletNode)
	wn.id = id
	wn.weight = 1
	tmp := len(initialScript)
	tmp -= 1024
	for tmp > 0 {
		wn.weight += 1
		tmp -= 4096
	}
	if q.walletRoot.weight+wn.weight > AtomsPerQuorum {
		err = errors.New("insufficient atoms in quorum")
		return
	}
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(Wallet)
	nw.id = id
	nw.Balance = balance
	nw.script = initialScript
	q.SaveWallet(nw)

	w.Balance.Subtract(balance)

	return
}

// "Cheat" function for initializing a bootstrap wallet
func (q *Quorum) CreateBootstrapWallet(id WalletID, Balance Balance, initialScript []byte) {
	// check if the new wallet already exists
	wn := q.retrieve(id)
	if wn != nil {
		panic("bootstrap wallet already exists")
	}

	// create a wallet node to insert into the walletTree
	wn = new(walletNode)
	wn.id = id
	wn.weight = 1
	tmp := len(initialScript)
	tmp -= 1024
	for tmp > 0 {
		wn.weight += 1
		tmp -= 4096
	}
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(Wallet)
	nw.id = id
	nw.Balance = Balance
	nw.script = initialScript
	q.SaveWallet(nw)
}

func (q *Quorum) Send(w *Wallet, amount Balance, destID WalletID) (cost int, err error) {
	cost += 1
	if !w.Balance.Compare(amount) {
		err = errors.New("insufficient balance")
		return
	}
	cost += 2
	destWallet := q.LoadWallet(destID)
	if destWallet == nil {
		err = errors.New("destination wallet does not exist")
		return
	}

	cost += 3
	w.Balance.Subtract(amount)
	destWallet.Balance.Add(amount)
	q.SaveWallet(destWallet)
	return
}

// JoinSia is a request that a wallet can submit to make itself a sibling in
// the quorum.
//
// The input is a sibling, a wallet (have to make sure that the wallet used
// as input is the sponsoring wallet...)
//
// Currently, AddSibling tries to add the new sibling to the existing quorum
// and throws the sibling out if there's no space. Once quorums are
// communicating, the AddSibling routine will always succeed.
func (q *Quorum) AddSibling(w *Wallet, s *Sibling) (cost int) {
	println("adding new sibling")
	cost = 50
	for i := 0; i < QuorumSize; i++ {
		if q.siblings[i] == nil {
			s.index = byte(i)
			s.wallet = w.id
			q.siblings[i] = s
			println("placed hopeful at index", i)
			break
		}
	}
	return
}

// Every wallet has a single sector, which can be up to 2^16 atoms of 4kb each,
// or 32GB total with 0 redundancy. Wallets pay for the size of their sector.
func (q *Quorum) ResizeSector(w *Wallet, atoms byte, m byte) (cost int, weight int, err error) {
	cost += 3
	weightDelta := int(atoms)
	weightDelta -= int(w.sectorAtoms)
	if weightDelta == 0 {
		return
	}

	// update the weights in the wallet tree
	q.updateWeight(w.id, weightDelta)
	if q.walletRoot.weight > AtomsPerQuorum {
		q.updateWeight(w.id, -weightDelta)
		return
	}
	weight = weightDelta

	// derive the name of the file housing the sector, and truncate the file
	walletName := q.walletFilename(w.id)
	sectorName := walletName + ".sector"
	err = os.Truncate(sectorName, int64(atoms)*int64(AtomSize))
	if err != nil {
		panic(err)
	}
	file, err := os.Open(sectorName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// update the hash associated with the sector
	w.sectorHash = q.MerkleCollapse(file)

	return
}

// First sectors are allocated, and then changes are uploaded to them. This
// creates a change.
func (q *Quorum) ProposeUpload(w *Wallet, parentHash siacrypto.Hash, newHashSet [QuorumSize]siacrypto.Hash, atomsChanged uint16, confirmations byte, deadline uint32) (cost int, weight uint16, err error) {
	cost += 2

	// make sure the sector is allocated
	if w.sectorAtoms == 0 {
		err = errors.New("Sector is not allocated")
		return
	}

	// make sure that the confirmations value is a reasonable value
	if int(confirmations) > QuorumSize {
		err = errors.New("confirmations cannot be greater than quorum size")
		return
	}
	if confirmations < w.sectorM {
		err = errors.New("confirmations cannot be less than the value of 'm' for the given sector")
		return
	}

	// make sure the deadline is a reasonable value
	if deadline > MaxDeadline+q.height {
		err = errors.New("deadline is too far in the future")
		return
	}
	if deadline <= q.height {
		err = errors.New("deadline has already arrived")
		return
	}

	cost += 2
	// look up all of the open uploads on this sector, and compare their hashes
	// to the parent hash of this upload. As soon as one is found (potentially
	// starting directly from the existing hash), all remaining uploads are
	// truncated. There can only exist a single chain of potential uploads, all
	// other get defeated by precedence.
	sectorID := string(w.id.Bytes())
	if parentHash == w.sectorHash {
		// clear all existing uploads
		q.clearUploads(sectorID, 0)
	} else {
		var i int
		for i = 0; i < len(q.uploads[sectorID]); i++ {
			if parentHash == q.uploads[sectorID][i].hash {
				break
			}
		}

		if i == len(q.uploads[sectorID]) {
			err = errors.New("upload has invalid parent hash")
			return
		}
		q.clearUploads(sectorID, i)
	}

	var uploadHash siacrypto.Hash
	for i := range newHashSet {
		uploadHash = siacrypto.CalculateHash(append(uploadHash[:], newHashSet[i][:]...))
	}
	u := upload{
		sectorID:              sectorID,
		requiredConfirmations: confirmations,
		hashSet:               newHashSet,
		hash:                  uploadHash,
		weight:                atomsChanged,
		deadline:              deadline,
	}

	cost += int((deadline - q.height) * uint32(atomsChanged+1) * q.storagePrice) // also need to add in the growth restraints
	weight = atomsChanged
	q.uploads[sectorID] = append(q.uploads[sectorID], &u)
	q.updateWeight(w.id, int(atomsChanged))
	q.insertEvent(&u)
	return
}
