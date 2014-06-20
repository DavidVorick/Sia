package quorum

import (
	"errors"
	"fmt"
	"os"
	"siaencoding"
)

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

func (q *Quorum) AllocateSector(w *Wallet, sector byte, atoms byte, m byte) (cost int) {
	weightDelta := int(atoms)
	weightDelta -= int(w.sectorOverview[sector].atoms)

	// derive the name of the file housing the sector
	walletName := q.walletFilename(w.id)
	sectorSlice := make([]byte, 1)
	sectorSlice[0] = sector
	sectorName := walletName + ".sector" + siaencoding.EncFilename(sectorSlice)

	// delete old file associated with the sector
	file, err := os.Create(sectorName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// write all the bytes to here
	emptySlice := make([]byte, int(atoms)*AtomSize)
	n, err := file.Write(emptySlice)
	if n != int(atoms)*AtomSize || err != nil {
		panic(err)
	}

	w.sectorOverview[sector].m = m
	w.sectorOverview[sector].atoms = atoms

	// update the weights in the wallet tree

	return
}

func (q *Quorum) ProposeUpload(w *Wallet, sector byte, atoms [256]bool, priority uint32, confirmations byte, deadline uint32) (cost int, err error) {
	// count the number of allocated atoms necessary for this upload
	cost += 2
	var i byte
	for i = 255; i >= 0; i-- {
		if atoms[i] == true {
			break
		}
	}

	// make sure the sector is allocated
	if w.sectorOverview[sector].atoms < i {
		err = errors.New("insufficient number of atoms allocated for Proposed Upload")
		return
	}

	// make sure that the confirmations value is a reasonable value
	if int(confirmations) > QuorumSize {
		err = errors.New("confirmations cannot be greater than quorum size")
		return
	}
	if confirmations < w.sectorOverview[sector].m {
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

	// make sure that all selected atoms are available to the sector
	cost += 2
	sectorString := fmt.Sprintf("%s%s", w.id.Bytes(), sector)
	u := upload{
		atoms:    atoms,
		priority: priority,
		deadline: deadline,
	}
	if q.uploads[sectorString] != nil {
		// check that there are no conflicting higher prioirty uploads in progress
		for i := range q.uploads[sectorString] {
			if q.uploads[sectorString][i].priority > priority {
				for j := range q.uploads[sectorString][i].atoms {
					if q.uploads[sectorString][i].atoms[j] == true && atoms[j] == true {
						err = errors.New("upload request is blocked by a higher priority upload in progress")
						return
					}
				}
			}
		}

		// overwrite any conflicting uploads of equal or lesser priority
		for i := range q.uploads[sectorString] {
			if q.uploads[sectorString][i].priority <= priority {
				for j := range q.uploads[sectorString][i].atoms {
					if q.uploads[sectorString][i].atoms[j] == true && atoms[j] == true {
						// remove the upload from the slice
						copy(q.uploads[sectorString][i-1:], q.uploads[sectorString][i+1:])
						q.uploads[sectorString] = q.uploads[sectorString][:len(q.uploads[sectorString])-1]

						// signal to the participant that the in-progress upload should be rejected
						// ...?
					}
				}
			}
		}

		// append the request to the uploads
		q.uploads[sectorString] = append(q.uploads[sectorString], &u)
	} else {
		q.uploads[sectorString] = make([]*upload, 1)
		q.uploads[sectorString][0] = &u
	}

	// add the upload to the eventList
	q.insertEvent(&u)
	return
}
