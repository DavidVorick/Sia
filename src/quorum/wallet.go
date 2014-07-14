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

// Takes an individual wallet and marshals it into something that can be sent
// over a wire.
func (w *Wallet) Marshal() (b []byte, err error) {
	b, err = siaencoding.Marshal(w)
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
