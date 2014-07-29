package state

import (
	"fmt"
	"math"
	"os"
	"siaencoding"
)

const (
	WalletIDSize         = 8
	walletAtomMultiplier = 3
)

type WalletID uint64

type Wallet struct {
	ID             WalletID
	Balance        Balance
	SectorSettings SectorSettings
	Script         []byte
}

func (id WalletID) Bytes() []byte {
	return siaencoding.EncUint64(uint64(id))
}

// Weight calculates and returns the weight of a wallet.
func (w Wallet) Weight() (weight uint32) {
	// Count the number of atoms used by the script.
	walletByteCount := float64(len(w.Script))
	walletAtomCount := walletByteCount / float64(AtomSize)
	walletAtomCount = math.Ceil(walletAtomCount)

	// Add an additional atom for the wallet itself.
	walletAtomCount += 1

	// Multiply script and wallet weight by the walletAtomMultiplier to account
	// for the snapshots that the wallet needs to reside in.
	walletAtomCount *= walletAtomMultiplier

	// Add non-replicated weight according to the size of the wallet sector.
	walletAtomCount += float64(w.SectorSettings.Atoms)
	walletAtomCount += float64(w.SectorSettings.UploadAtoms)
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
	wn.weight = int(w.Weight())
	s.insertWalletNode(wn)

	s.SaveWallet(w)
	return
}

// LoadWallet checks the wallettree for existence of the wallet, and then loads
// the wallet from disk if the wallet exists.
func (s *State) LoadWallet(id WalletID) (w Wallet, err error) {
	// Check that the wallet is in the wallettree.
	wn := s.walletNode(id)
	if wn == nil {
		err = fmt.Errorf("LoadWallet: no wallet of that id exists.")
		return
	}

	// Fetch the wallet filename and open the file.
	walletFilename := s.walletFilename(id)
	file, err := os.Open(walletFilename)
	if err != nil {
		panic(err)
	}

	// Fetch the size of the wallet from disk.
	walletLengthBytes := make([]byte, 4)
	_, err = file.Read(walletLengthBytes)
	if err != nil {
		panic(err)
	}
	walletLength := siaencoding.DecUint32(walletLengthBytes)

	// Fetch the wallet from disk and decode it.
	walletBytes := make([]byte, walletLength)
	_, err = file.Read(walletBytes)
	if err != nil {
		panic(err)
	}
	err = siaencoding.Unmarshal(walletBytes, &w)
	if err != nil {
		panic(err)
	}
	return
}

// SaveWallet takes a wallet object and updates the corresponding walletNode,
// and then saves the wallet to disk.
func (s *State) SaveWallet(w Wallet) (err error) {
	// Check that the wallet is in the wallettree.
	wn := s.walletNode(w.ID)
	if wn == nil {
		err = fmt.Errorf("SaveWallet: no wallet of that id exists.")
		return
	}
	weightDelta := int(w.Weight()) - wn.nodeWeight()
	// Ideally, this would never be triggered. Instead, careful resource
	// management in the quorum would prevent a too-heavy wallet from ever
	// getting this far through the insert process.
	if s.walletRoot.weight+weightDelta > AtomsPerQuorum {
		err = fmt.Errorf("SaveWallet: wallet is too heavy to fit in the quorum.")
		return
	}
	s.updateWeight(w.ID, weightDelta)

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
