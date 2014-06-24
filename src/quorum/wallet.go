package quorum

import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
)

const (
	walletBaseSize       = siacrypto.HashSize + 16 + 2 + 1 + siacrypto.HashSize + 2
	walletAtomMultiplier = 3
)

// the default script for all wallets; simply transfers control to input
// eventually this will be modified to verify a public key before executing
var genesisScript = []byte{0x28}

type Wallet struct {
	id WalletID

	walletHash siacrypto.Hash
	Balance    Balance

	sectorAtoms uint16
	sectorM     byte
	sectorHash  siacrypto.Hash

	scriptAtoms uint16
	script      []byte
}

func (w *Wallet) Script() []byte {
	return w.script
}

// takes a walletID and derives the filename from the quorum. Eventually, this
// function should also verify that the id is located within the quorum.
func (q *Quorum) walletFilename(id WalletID) (s string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siaencoding.EncFilename(suffixBytes)
	s = q.walletPrefix + suffix
	return
}

func (q *Quorum) walletString(id WalletID) (s string) {
	w := q.LoadWallet(id)
	if w == nil {
		return "\t\t\tError! Don't have wallet!\n"
		return
	}
	s += fmt.Sprintf("\t\t\tUpper Balance: %v\n", w.Balance.upperBalance)
	s += fmt.Sprintf("\t\t\tLower Balance: %v\n", w.Balance.lowerBalance)
	s += fmt.Sprintf("\t\t\tSector Atoms: %v\n", w.sectorAtoms)
	s += fmt.Sprintf("\t\t\tSector M: %v\n", w.sectorM)
	s += fmt.Sprintf("\t\t\tSector Hash: %v\n", w.sectorHash[:6])
	s += fmt.Sprintf("\t\t\tScript Atoms: %v\n", w.scriptAtoms)
	s += fmt.Sprintf("\t\t\tScript Length: %v\n", len(w.script))
	return
}

// takes a wallet and converts it to a byte slice. Considering changing the
// name to GobEncode but not sure if that's needed. The hash is calculated
// after encoding the rest of the wallet - it is this function alone that is
// responsible for creating a hash that verifies the integrity of the wallet.
func (w *Wallet) GobEncode() (b []byte, err error) {
	if w == nil {
		err = fmt.Errorf("Cannot encode nil wallet")
		return
	}

	b = make([]byte, walletBaseSize+len(w.script))

	// leave room for Hash, encode balance and scriptAtoms
	offset := siacrypto.HashSize
	balanceBytes, err := w.Balance.GobEncode()
	if err != nil {
		return
	}
	copy(b[offset:], balanceBytes)
	offset += 16

	// encode sector information
	sectorAtomsBytes := siaencoding.EncUint16(w.sectorAtoms)
	copy(b[offset:], sectorAtomsBytes)
	offset += 2
	b[offset] = w.sectorM
	offset += 1
	copy(b[offset:], w.sectorHash[:])
	offset += siacrypto.HashSize

	// encode script information
	scriptAtomsBytes := siaencoding.EncUint16(w.scriptAtoms)
	copy(b[offset:], scriptAtomsBytes)
	offset += 2
	copy(b[offset:], w.script)

	// calculate hash and place at beginning
	hash := siacrypto.CalculateHash(b[siacrypto.HashSize:])
	copy(b, hash[:])

	return
}

// upon decoding, the hash is checked to make sure that wallet integrity was
// maintained.
func (w *Wallet) GobDecode(b []byte) (err error) {
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

	// decode balance
	err = w.Balance.GobDecode(b[offset : offset+16])
	if err != nil {
		return
	}
	offset += 16

	// decode sector information
	w.sectorAtoms = siaencoding.DecUint16(b[offset : offset+2])
	offset += 2
	w.sectorM = b[offset]
	offset += 1
	copy(w.sectorHash[:], b[offset:])
	offset += siacrypto.HashSize

	// decode script informaiton
	w.scriptAtoms = siaencoding.DecUint16(b[offset : offset+2])
	offset += 2
	w.script = b[offset:]
	return
}

func (q *Quorum) InsertWallet(encodedWallet []byte, id WalletID) (err error) {
	w := new(Wallet)
	err = w.GobDecode(encodedWallet)
	if err != nil {
		return
	}
	w.id = id

	wn := q.retrieve(id)
	if wn != nil {
		err = fmt.Errorf("InsertWallet: wallet of that id already exists in quorum")
		return
	}

	weight := walletAtomMultiplier + w.scriptAtoms*walletAtomMultiplier
	weight += w.sectorAtoms

	wn = new(walletNode)
	wn.id = id
	wn.weight = int(weight)
	q.insert(wn)

	q.SaveWallet(w)
	return
}

func (q *Quorum) LoadWallet(id WalletID) (w *Wallet) {
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

	w = new(Wallet)
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
func (q *Quorum) SaveWallet(w *Wallet) {
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
