package quorum

import (
	"encoding/base64"
	"fmt"
	"os"
	"siacrypto"
	"strings"
)

const (
	walletBaseSize      = siacrypto.HashSize + 16 + 2 + 256*8
	CreateWalletMaxCost = 8
)

// the default script for all wallets; simply transfers control to input
// eventually this will be modified to verify a public key before executing
var genesisScript = []byte{0x28}

type sectorHeader struct {
	crc   [6]byte
	m     byte
	atoms byte
}

type wallet struct {
	id WalletID

	walletHash     siacrypto.Hash // a hash of the encoded wallet
	balance        Balance
	sectorOverview [256]sectorHeader
	script         []byte
}

func (q *Quorum) walletString(id WalletID) (s string) {
	w := q.loadWallet(id)
	s += fmt.Sprintf("\t\t\tUpper Balance: %v\n", w.balance.upperBalance)
	s += fmt.Sprintf("\t\t\tLower Balance: %v\n", w.balance.lowerBalance)
	s += fmt.Sprintf("\t\t\tScript Length: %v\n", len(w.script))

	// calculate the number of sectors that have been allocated
	allocatedSectors := 0
	for _, sectorHeader := range w.sectorOverview {
		if sectorHeader.atoms != 0 {
			allocatedSectors += 1
		}
	}
	s += fmt.Sprintf("\t\t\tAllocated Sectors: %v\n", allocatedSectors)
	return
}

// takes a wallet and converts it to a byte slice. Considering changing the
// name to GobEncode but not sure if that's needed. The hash is calculated
// after encoding the rest of the wallet - it is this function alone that is
// responsible for creating a hash that verifies the integrity of the wallet.
func (w *wallet) GobEncode() (b []byte, err error) {
	if w == nil {
		err = fmt.Errorf("Cannot encode nil wallet")
		return
	}

	b = make([]byte, walletBaseSize+len(w.script))

	// leave room for Hash, encode balance and scriptAtoms
	offset := siacrypto.HashSize
	balanceBytes, err := w.balance.GobEncode()
	if err != nil {
		return
	}
	copy(b[offset:], balanceBytes)
	offset += 16

	// encode sectorOverivew
	for i, sector := range w.sectorOverview {
		copy(b[offset+i*8:], sector.crc[:])
		b[offset+i*8+6] = sector.m
		b[offset+i*8+7] = sector.atoms
	}
	offset += 8 * len(w.sectorOverview)

	// encode script
	copy(b[offset:], w.script)

	// calculate hash and place at beginning
	hash := siacrypto.CalculateHash(b[siacrypto.HashSize:])
	copy(b, hash[:])

	return
}

// upon decoding, the hash is checked to make sure that wallet integrity was
// maintained.
func (w *wallet) GobDecode(b []byte) (err error) {
	if w == nil {
		err = fmt.Errorf("Cannot decode into nil wallet")
		return
	}

	// verify the integrity of the wallet
	copy(w.walletHash[:], b)
	expectedHash := siacrypto.CalculateHash(b[siacrypto.HashSize:])
	if expectedHash != w.walletHash {
		err = fmt.Errorf("Wallet Gob Decode: hash does not match wallet!")
		return
	}
	offset := siacrypto.HashSize

	err = w.balance.GobDecode(b[offset : offset+16])
	if err != nil {
		return
	}
	offset += 16

	for i := range w.sectorOverview {
		copy(w.sectorOverview[i].crc[:], b[offset+i*8:offset+i*8+6])
		w.sectorOverview[i].m = b[offset+i*8+6]
		w.sectorOverview[i].atoms = b[offset+i*8+7]
	}
	offset += 8 * len(w.sectorOverview)

	copy(w.script, b[offset:])
	return
}

func (q *Quorum) LoadWallet(encodedWallet []byte, id WalletID) (err error) {
	w := new(wallet)
	err = w.GobDecode(encodedWallet)
	if err != nil {
		return
	}

	wn := q.retrieve(id)
	if wn != nil {
		err = fmt.Errorf("LoadWallet: wallet of that id already exists in quorum")
		return
	}

	weight := 1
	if len(w.script) > 1024 {
		tmp := len(w.script) - 1024
		for tmp > 0 {
			weight += 1
			tmp -= 4096
		}
	}

	for _, sector := range w.sectorOverview {
		weight += int(sector.atoms)
	}

	wn = new(walletNode)
	wn.id = id
	wn.weight = int(weight)
	q.insert(wn)

	q.saveWallet(w)
	return
}

// takes a walletID and derives the filename from the quorum. Eventually, this
// function should also verify that the id is located within the quorum.
func (q *Quorum) walletFilename(id WalletID) (s string) {
	// Turn the id into a suffix that will follow the quorum prefix
	walletSuffix := make([]byte, 8)
	for i := 0; i < 8; i++ {
		walletSuffix[i] = byte(id)
		id = id >> 8
	}
	safeSuffix := base64.StdEncoding.EncodeToString(walletSuffix)
	safeSuffix = strings.Replace(safeSuffix, "/", "_", -1)
	s = q.walletPrefix + safeSuffix
	return
}

func (q *Quorum) loadWallet(id WalletID) (w *wallet) {
	walletFilename := q.walletFilename(id)
	file, err := os.Open(walletFilename)
	if err != nil {
		return nil
	}

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	b := make([]byte, fileInfo.Size())

	_, err = file.Read(b)
	if err != nil {
		panic(err)
	}

	err = w.GobDecode(b)
	if err != nil {
		panic(err)
	}
	w.id = id
	return
}

// takes a wallet as input, then uses the quorum prefix plus the wallet id to
// determine the filename for the wallet. Then it writes a 4kb block of data to
// the wallet file and saves it to disk.
func (q *Quorum) saveWallet(w *wallet) {
	walletFilename := q.walletFilename(w.id)
	file, err := os.Create(walletFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	walletBytes, err := w.GobEncode()
	if err != nil {
		panic(err)
	}
	_, err = file.Write(walletBytes[:])
	if err != nil {
		panic(err)
	}
}

// CreateWallet takes an id, a balance, a number of script atom, and an initial
// script and uses those to create a new wallet that gets stored in stable
// memory. If a wallet of that id already exists then the process aborts.
func (q *Quorum) CreateWallet(w *wallet, id WalletID, balance Balance, scriptAtoms uint16, initialScript []byte) (cost int) {
	cost += 1
	if !w.balance.Compare(balance) {
		return
	}

	// check if the new wallet already exists
	cost += 2
	wn := q.retrieve(id)
	if wn != nil {
		return
	}

	// create a wallet node to insert into the walletTree
	cost += 5
	wn = new(walletNode)
	wn.id = id
	wn.weight = int(1 + scriptAtoms)
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(wallet)
	nw.id = id
	nw.balance = balance
	copy(nw.script, initialScript)
	q.saveWallet(nw)

	w.balance.Subtract(balance)

	return
}
