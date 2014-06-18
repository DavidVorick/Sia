package quorum

import (
	"errors"
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
