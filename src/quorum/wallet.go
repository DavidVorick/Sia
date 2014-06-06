package quorum

import (
	"encoding/base64"
	"fmt"
	"os"
	"siacrypto"
	"strings"
)

const (
	scriptPrimerSize = 3534
)

type WalletID uint64
type walletHandle [8]byte

type sectorHeader struct {
	m        byte
	numAtoms byte
}

// A 4kb block of data containing everything about a wallet.
type wallet struct {
	walletHash     siacrypto.TruncatedHash //hash of the 4064 bytes
	upperBalance   uint64
	lowerBalance   uint64
	scriptAtoms    uint16
	sectorOverview [256]sectorHeader
	// above section comes to 564 bytes

	scriptPrimer [scriptPrimerSize]byte
	// all together, that's 4kb
}

// converts a WalletID to a walletHandle
func (id WalletID) handle() (b walletHandle) {
	for i := 0; i < 8; i++ {
		b[i] = byte(id)
		id = id >> 8
	}
	return
}

// converts a walletHandle to a WalletID
func (h walletHandle) id() (b WalletID) {
	var a uint64
	for i := 7; i > 0; i-- {
		a += uint64(h[i])
		a = a << 8
	}
	a += uint64(h[0])

	b = WalletID(a)
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

func (q *Quorum) CreateWallet(id WalletID, upperBalance uint64, lowerBalance uint64, scriptAtoms uint16, initialScript []byte) (err error) {
	// check if the wallet already exists
	walletNode := q.retrieve(id)
	if walletNode != nil {
		err = fmt.Errorf("CreateWallet: wallet of that id already exists in quorum.")
		return
	}

	// create a wallet node to insert into the walletTree
	wn := new(walletNode)
	wn.id = id
	wn.weight = 1+scriptAtoms
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	var w wallet
	w.upperBalance = upperBalance
	w.lowerBalance = lowerBalance
	w.scriptAtoms = scriptAtoms
	copy(w.scriptPrimer[:], initialScript)

	// derive the base64 filename for the wallet, converting slashes for filesystem compatibility
	walletHandleBytes := [8]byte(id.handle())
	walletFilenameBytes := append([]byte(q.walletPrefix), walletHandleBytes[:]...)
	walletFilename := base64.StdEncoding.EncodeToString(walletFilenameBytes)
	walletFilename = strings.Replace(walletFilename, "/", "_", -1)

	// open the file and write the wallets
	file, err := os.Create(walletFilename)
	if err != nil {
		return
	}
	defer file.Close()
	walletBytes := w.bytes()
	_, err = file.Write(walletBytes[:])
	if err != nil {
		return
	}

	// Allocate script atoms
	if scriptAtoms != 0 {
		scriptBytes := make([]byte, 4096*scriptAtoms)
		if len(scriptBytes) > scriptPrimerSize {
			copy(scriptBytes, initialScript[scriptPrimerSize:])
		}
		_, err = file.Write(scriptBytes)
		if err != nil {
			return
		}
	}

	return
}
