package quorum

import (
	"bytes"
	"encoding/gob"
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
	if w.Balance.Compare(balance) < 0 {
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
	wn.weight = walletAtomMultiplier
	tmp := len(initialScript)
	tmp -= 1024
	var scriptAtoms uint16
	for tmp > 0 {
		wn.weight += walletAtomMultiplier
		tmp -= 4096
		scriptAtoms += 1
	}
	if q.walletRoot.weight+wn.weight > AtomsPerQuorum {
		err = errors.New("insufficient atoms in quorum")
		return
	}
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(Wallet)
	nw.ID = id
	nw.Balance = balance
	nw.Script = initialScript
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
	wn.weight = walletAtomMultiplier
	tmp := len(initialScript)
	tmp -= 1024
	for tmp > 0 {
		wn.weight += 1
		tmp -= 4096
	}
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(Wallet)
	nw.ID = id
	nw.Balance = Balance
	nw.Script = initialScript
	q.SaveWallet(nw)
}

func (q *Quorum) Send(w *Wallet, amount Balance, destID WalletID) (cost int, err error) {
	cost += 1
	if w.Balance.Compare(amount) < 0 {
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

// Currently, AddSibling tries to add the new sibling to the existing quorum
// and throws the sibling out if there's no space. Once quorums are
// communicating, the AddSibling routine will always succeed.
func (q *Quorum) AddSibling(w *Wallet, s *Sibling) (cost int) {
	println("adding new sibling")
	cost = 50
	for i := byte(0); i < QuorumSize; i++ {
		if q.siblings[i] == nil {
			s.index = byte(i)
			s.wallet = w.ID
			q.siblings[i] = s
			println("placed hopeful at index", i)
			break
		}
	}
	return
}

// Every wallet has a single sector, which can be up to 2^16 atoms of 4kb each,
// or 32GB total with 0 redundancy. Wallets pay for the size of their sector.
func (q *Quorum) ResizeSectorErase(w *Wallet, atoms uint16, k byte) (cost int, weight int, err error) {
	cost += 3
	weightDelta := int(atoms)
	// weightDelta -= int(w.sectorAtoms)
	if weightDelta == 0 {
		return
	}

	// update the weights in the wallet tree
	q.updateWeight(w.ID, weightDelta)
	if q.walletRoot.weight > AtomsPerQuorum {
		q.updateWeight(w.ID, -weightDelta)
		return
	}
	weight = weightDelta

	// remove the file and return if the sector has been resized to length 0
	walletName := q.walletFilename(w.ID)
	sectorName := walletName + ".sector"
	if atoms == 0 {
		os.Remove(sectorName)
		return
	}

	// derive the name of the file housing the sector, and truncate the file
	file, err := os.Create(sectorName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// extend the file to being to proper length
	err = file.Truncate(int64(atoms) * int64(AtomSize))
	if err != nil {
		panic(err)
	}

	// update the hash associated with the sector
	_, err = file.Seek(int64(AtomSize), 0) // first atom contains hash information
	if err != nil {
		panic(err)
	}
	zeroMerkle := MerkleCollapse(file)

	// build the first atom of the file to contain all of the hashes
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	for i := byte(0); i < QuorumSize; i++ {
		_, err := file.Write(zeroMerkle[:])
		if err != nil {
			panic(err)
		}
	}

	// get the hash of the first atom as the sector hash
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	firstAtom := make([]byte, AtomSize)
	_, err = file.Read(firstAtom)
	if err != nil {
		panic(err)
	}
	// w.sectorAtoms = atoms
	// w.sectorM = k
	// w.sectorHash = siacrypto.CalculateHash(firstAtom)

	return
}

type UploadArgs struct {
	ParentHash    siacrypto.Hash
	NewHashSet    [QuorumSize]siacrypto.Hash
	AtomsChanged  uint16
	Confirmations byte
	Deadline      uint32
}

func (ua *UploadArgs) GobEncode() (gobUA []byte, err error) {
	if ua == nil {
		return
	}
	b := new(bytes.Buffer)
	encoder := gob.NewEncoder(b)

	err = encoder.Encode(ua.ParentHash)
	if err != nil {
		return
	}
	err = encoder.Encode(ua.NewHashSet)
	if err != nil {
		return
	}
	err = encoder.Encode(ua.AtomsChanged)
	if err != nil {
		return
	}
	err = encoder.Encode(ua.Confirmations)
	if err != nil {
		return
	}
	err = encoder.Encode(ua.Deadline)
	if err != nil {
		return
	}

	gobUA = b.Bytes()
	return
}

func (ua *UploadArgs) GobDecode(gobUA []byte) (err error) {
	b := bytes.NewBuffer(gobUA)
	decoder := gob.NewDecoder(b)

	err = decoder.Decode(&ua.ParentHash)
	if err != nil {
		return
	}
	err = decoder.Decode(&ua.NewHashSet)
	if err != nil {
		return
	}
	err = decoder.Decode(&ua.AtomsChanged)
	if err != nil {
		return
	}
	err = decoder.Decode(&ua.Confirmations)
	if err != nil {
		return
	}
	err = decoder.Decode(&ua.Deadline)
	if err != nil {
		return
	}
	return
}

// First sectors are allocated, and then changes are uploaded to them. This
// creates a change.
func (q *Quorum) ProposeUpload(w *Wallet, parentHash siacrypto.Hash, newHashSet [QuorumSize]siacrypto.Hash, atomsChanged uint16, confirmations byte, deadline uint32) (cost int, weight uint16, err error) {
	cost += 2

	// make sure the sector is allocated
	//if w.sectorAtoms == 0 {
	//		err = errors.New("Sector is not allocated")
	//		return
	//	}

	// make sure that the confirmations value is a reasonable value
	if confirmations > QuorumSize {
		err = errors.New("confirmations cannot be greater than quorum size")
		return
	}
	//if confirmations < w.sectorM {
	//	err = errors.New("confirmations cannot be less than the value of 'm' for the given sector")
	//		return
	//	}

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
	/* if parentHash == w.sectorHash {
		// clear all existing uploads
		q.clearUploads(w.id, 0)
	} else {
		var i int
		for i = 0; i < len(q.uploads[w.id]); i++ {
			if parentHash == q.uploads[w.id][i].hash {
				break
			}
		}

		if i == len(q.uploads[w.id]) {
			err = errors.New("upload has invalid parent hash")
			return
		}
		q.clearUploads(w.id, i)
	}*/

	uploadHash := SectorHash(newHashSet)
	u := upload{
		id: w.ID,
		requiredConfirmations: confirmations,
		hashSet:               newHashSet,
		hash:                  uploadHash,
		weight:                atomsChanged,
		deadline:              deadline,
	}

	weight = atomsChanged
	if q.uploads[w.ID] == nil {
		q.uploads[w.ID] = make([]*upload, 0)
	}
	q.uploads[w.ID] = append(q.uploads[w.ID], &u)
	q.updateWeight(w.ID, int(atomsChanged))
	q.insertEvent(&u)
	return
}
