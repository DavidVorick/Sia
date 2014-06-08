package quorum

import (
	"encoding/base64"
	"fmt"
	"os"
	"siacrypto"
	"strings"
)

const (
	scriptPrimerSize = 1024
)

// the default script for all wallets; simply transfers control to input
// eventually this will be modified to verify a public key before executing
var genesisScript = []byte{0x28}

type WalletID uint64

type sectorHeader struct {
	// 6 byte CRC will go here
	m        byte
	numAtoms byte
}

// A 4kb block of data containing everything about a wallet.
type wallet struct {
	id WalletID // not saved to disk

	walletHash     siacrypto.TruncatedHash //hash of the 4064 bytes, need to be specification-based
	upperBalance   uint64
	lowerBalance   uint64
	scriptAtoms    uint16
	sectorOverview [256]sectorHeader
	// above section comes to 566 bytes

	scriptPrimer [scriptPrimerSize]byte
}

func (q *Quorum) walletString(id WalletID) (s string) {
	w := q.loadWallet(id)
	s += fmt.Sprintf("\t\t\tUpper Balance: %v\n", w.upperBalance)
	s += fmt.Sprintf("\t\t\tLower Balance: %v\n", w.lowerBalance)
	s += fmt.Sprintf("\t\t\tScript Atoms: %v\n", w.scriptAtoms)

	// calculate the number of sectors that have been allocated
	allocatedSectors := 0
	for _, sectorHeader := range w.sectorOverview {
		if sectorHeader.numAtoms != 0 {
			allocatedSectors += 1
		}
	}
	s += fmt.Sprintf("\t\t\tAllocated Sectors: %v\n", allocatedSectors)
	return
}

// Takes a wallet as input and produces a [4096]byte as an output, like an
// encoding but explicity designed to be stored on a disk.
func (w *wallet) bytes() (b *[4096]byte) {
	b = new([4096]byte)

	// The hash is calculated before returning the bytes. This allows whomever is
	// using the wallet to update the values without needing to update the hash
	// every time. The hash is only calculated when the wallet is being converted
	// to a string of bytes.
	offset := siacrypto.TruncatedHashSize

	tmp := w.upperBalance
	for i := 0; i < 8; i++ {
		b[offset+i] = byte(tmp)
		tmp = tmp >> 8
	}
	offset += 8

	tmp = w.lowerBalance
	for i := 0; i < 8; i++ {
		b[offset+i] = byte(tmp)
		tmp = tmp >> 8
	}
	offset += 8

	tmp16 := w.scriptAtoms
	for i := 0; i < 2; i++ {
		b[offset+i] = byte(tmp16)
		tmp16 = tmp16 >> 8
	}
	offset += 2

	for i, sector := range w.sectorOverview {
		b[offset+i*2] = sector.m
		b[offset+i*2+1] = sector.numAtoms
	}
	offset += 2 * len(w.sectorOverview)

	copy(b[offset:], w.scriptPrimer[:])

	hash, err := siacrypto.CalculateTruncatedHash(b[:][32:])
	if err != nil {
		return nil
	}
	copy(b[:], hash[:])
	return
}

func fillWallet(b *[4096]byte) (w *wallet) {
	w = new(wallet)
	copy(w.walletHash[:], b[:])
	// do an integrity check of the wallet, return nil if there are errors during
	// the check
	expectedHash, err := siacrypto.CalculateTruncatedHash(b[:][32:])
	if err != nil || expectedHash != w.walletHash {
		// if err != nil, there should probably a more severe thing.  maybe
		// CalculateTruncatedHash shouldn't return an error at all, and instead
		// call panic or some extreme logging function.
		return nil
	}
	offset := siacrypto.TruncatedHashSize

	for i := 7; i > 0; i-- {
		w.upperBalance += uint64(b[offset+i])
		w.upperBalance = w.upperBalance << 8
	}
	w.upperBalance += uint64(b[offset])
	offset += 8

	for i := 7; i > 0; i-- {
		w.lowerBalance += uint64(b[offset+i])
		w.lowerBalance = w.lowerBalance << 8
	}
	w.lowerBalance += uint64(b[offset])
	offset += 8

	w.scriptAtoms += uint16(b[offset+1])
	w.scriptAtoms = w.scriptAtoms << 8
	w.scriptAtoms += uint16(b[offset])
	offset += 2

	for i := range w.sectorOverview {
		w.sectorOverview[i].m = b[offset+i*2]
		w.sectorOverview[i].numAtoms = b[offset+i*2+1]
	}
	offset += 2 * len(w.sectorOverview)

	copy(w.scriptPrimer[:], b[offset:])
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

func (q *Quorum) loadWallet(id WalletID) *wallet {
	walletFilename := q.walletFilename(id)
	file, err := os.Open(walletFilename)
	if err != nil {
		panic(err)
	}

	var b [4096]byte
	n, err := file.Read(b[:])
	if err != nil {
		panic(err)
	}
	if n != 4096 {
		return nil
	}

	return fillWallet(&b)
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
	walletBytes := w.bytes() // w.bytes() takes care of the hashing
	_, err = file.Write(walletBytes[:])
	if err != nil {
		panic(err)
	}
}

func (q *Quorum) LoadScript(id WalletID) []byte {
	w := q.loadWallet(id)
	if w.scriptAtoms == 0 {
		return w.scriptPrimer[:]
	}

	// load script block
	scriptBody := make([]byte, 4096*w.scriptAtoms)
	scriptName := q.walletFilename(id)
	scriptName += "script"
	file, err := os.Open(scriptName)
	if err != nil {
		panic(err) // this will really result in fetching the data
	}
	_, err = file.Read(scriptBody)
	if err != nil {
		panic(err)
	}

	return append(w.scriptPrimer[:], scriptBody...)
}

func (q *Quorum) SaveScript(id WalletID, scriptBlock []byte) {
	// check that scriptBlock is the correct size
	w := q.loadWallet(id)
	if len(scriptBlock) != int(1024+4096*w.scriptAtoms) {
		return
	}
	copy(w.scriptPrimer[:], scriptBlock)
	q.saveWallet(w)

	if len(scriptBlock) <= 1024 {
		return
	}

	scriptFilename := q.walletFilename(id)
	scriptFilename += ".script"
	file, err := os.Create(scriptFilename)
	defer file.Close()
	if err != nil {
		return
	}

	_, err = file.Write(scriptBlock[scriptPrimerSize:])
	if err != nil {
		return
	}

	// Write the other script atoms too
}

// CreateWallet takes an id, a balance, a number of script atom, and an initial
// script and uses those to create a new wallet that gets stored in stable
// memory. If a wallet of that id already exists then the process aborts.
func (q *Quorum) CreateWallet(id WalletID, upperBalance uint64, lowerBalance uint64, scriptAtoms uint16, initialScript []byte) (err error) {
	// check if the wallet already exists
	wn := q.retrieve(id)
	if wn != nil {
		err = fmt.Errorf("CreateWallet: wallet of that id already exists in quorum.")
		return
	}

	// create a wallet node to insert into the walletTree
	wn = new(walletNode)
	wn.id = id
	wn.weight = int(1 + scriptAtoms)
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	w := new(wallet)
	w.id = id
	w.upperBalance = upperBalance
	w.lowerBalance = lowerBalance
	w.scriptAtoms = scriptAtoms

	q.saveWallet(w)
	q.SaveScript(w.id, initialScript) // scripts are handled separately
	return
}
