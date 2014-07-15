package quorum

import (
	"fmt"
	"os"
	"siaencoding"
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

// Weight calculates and returns the weight of a wallet.
func (w Wallet) Weight() (weight uint32) {
	walletByteCount := float64(len(w.Script))
	walletAtomCount := walletByteCount / float64(AtomSize)
	walletAtomCount += 1
	walletAtomCount *= walletAtomMultiplier
	return uint32(walletAtomCount)
}

// InsertWallet takes a new wallet and inserts it into the wallet tree.
// InsertWallet returns an error if the wallet already exists within the state.
func (s *State) InsertWallet(w Wallet) (err error) {
	wn := s.walletNode(w.ID)
	if wn != nil {
		err = fmt.Errorf("InsertWallet: wallet of that id already exists in quorum")
		return
	}

	wn = new(walletNode)
	wn.id = w.ID
	s.insertWalletNode(wn)

	s.SaveWallet(w)
	return
}

func (s *State) RemoveWallet(id WalletID) {
	// Delete the file that contains the wallet on disk.
	filename := s.walletFilename(id)
	err := os.Remove(filename)
	if err != nil {
		panic(err)
	}

	// Delete from wallet tree.
	s.removeWalletNode(id)
}

// LoadWallet checks the disk for a saved wallet, and loads that wallet into
// memory. No changes to State are made.
func (s *State) LoadWallet(id WalletID) (w Wallet, err error) {
	// Fetch the wallet filename and open the file.
	walletFilename := s.walletFilename(id)
	file, err := os.Open(walletFilename)
	if err != nil {
		return
	}

	// Fetch the size of the wallet from disk.
	walletLengthBytes := make([]byte, 4)
	_, err = file.Read(walletLengthBytes)
	if err != nil {
		return
	}
	walletLength := siaencoding.DecUint32(walletLengthBytes)

	// Fetch the wallet from disk and decode it.
	walletBytes := make([]byte, walletLength)
	_, err = file.Read(walletBytes)
	if err != nil {
		return
	}
	err = siaencoding.Unmarshal(walletBytes, &w)
	if err != nil {
		return
	}
	return
}

// SaveWallet takes a wallet object and saves it to disk. SaveWallet does not
// update the corresponding wallet node or make any changes other than saving
// the wallet to disk.
func (s *State) SaveWallet(w Wallet) {
	// Fetch the wallet filename from the state object.
	walletFilename := s.walletFilename(w.ID)
	file, err := os.Create(walletFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Encode the wallet to a byte slice.
	walletBytes, err := siaencoding.Marshal(w)
	if err != nil {
		panic(err)
	}
	// Encode the length of the byte slice.
	lengthPrefix := siaencoding.EncUint32(uint32(len(walletBytes)))
	// Write the length prefix to the file.
	_, err = file.Write(lengthPrefix[:])
	if err != nil {
		panic(err)
	}
	// Write the wallet to the file.
	_, err = file.Write(walletBytes[:])
	if err != nil {
		panic(err)
	}
}
