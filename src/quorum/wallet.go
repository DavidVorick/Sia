package quorum

import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
	"siafiles"
)

const (
	walletAtomMultiplier = 3
)

// the default script for all wallets; simply transfers control to input
// eventually this will be modified to verify a public key before executing
var genesisScript = []byte{0x28}

type Wallet struct {
	ID             WalletID
	Balance        Balance
	SectorSettings SectorSettings
	Script         []byte
}

func (q *Quorum) MarshalWallet(id WalletID) (b []byte, err error) {
	// instead of marshalling the id, you have to fetch the wallet from the
	// wallet tree, load it off the disk or whatever, and then use it.

	b, err = siaencoding.Marshal(id)
	return
}

// takes a walletID and derives the filename from the quorum. Eventually, this
// function should also verify that the id is located within the quorum.
func (q *Quorum) walletFilename(id WalletID) (s string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	s = q.walletPrefix + suffix
	return
}

func (q *Quorum) walletString(id WalletID) (s string) {
	w := q.LoadWallet(id)
	if w == nil {
		return "\t\t\tError! Don't have wallet!\n"
		return
	}
	s += fmt.Sprintf("\t\t\tBalance: %v\n", siaencoding.DecUint128(w.Balance[:]))
	//s += fmt.Sprintf("\t\t\tSector Atoms: %v\n", w.sectorAtoms)
	//s += fmt.Sprintf("\t\t\tSector M: %v\n", w.sectorM)
	//s += fmt.Sprintf("\t\t\tSector Hash: %v\n", w.sectorHash[:6])
	//s += fmt.Sprintf("\t\t\tScript Atoms: %v\n", w.scriptAtoms)
	s += fmt.Sprintf("\t\t\tScript Length: %v\n", len(w.Script))
	return
}

func (q *Quorum) InsertWallet(encodedWallet []byte, id WalletID) (err error) {
	w := new(Wallet)
	err = w.GobDecode(encodedWallet)
	if err != nil {
		return
	}
	w.ID = id

	wn := q.retrieve(id)
	if wn != nil {
		err = fmt.Errorf("InsertWallet: wallet of that id already exists in quorum")
		return
	}

	//weight := walletAtomMultiplier + w.scriptAtoms*walletAtomMultiplier
	//weight += w.sectorAtoms

	wn = new(walletNode)
	wn.id = id
	//wn.weight = int(weight)
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
	w.ID = id
	return
}

// takes a wallet as input, then uses the quorum prefix plus the wallet id to
// determine the filename for the wallet. Then it writes a 4kb block of data to
// the wallet file and saves it to disk.
func (q *Quorum) SaveWallet(w *Wallet) {
	walletFilename := q.walletFilename(w.ID)
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
