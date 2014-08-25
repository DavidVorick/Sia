package state

import (
	"errors"
	"fmt"
	"os"

	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/siafiles"
)

const (
	// WalletIDSize is the size of a WalletID in bytes.
	WalletIDSize         = 8
	walletAtomMultiplier = 3
)

// A WalletID is a unique identifier that references a Wallet on the network.
type WalletID uint64

// A Wallet performs three important duties. It contains a Balance, allowing
// for transactions; a SectorSettings object which manages what storage is
// associated with the Wallet; and a Script, which can receive inputs and
// perform actions.
type Wallet struct {
	ID             WalletID
	Balance        Balance
	SectorSettings SectorSettings
	Script         []byte
	KnownScripts   map[string]ScriptInputEvent
}

// Bytes returns the WalletID as a byte slice.
func (id WalletID) Bytes() []byte {
	return siaencoding.EncUint64(uint64(id))
}

// TODO: add docstring
func (w Wallet) CompensationWeight() (weight uint32) {
	// Count the number of atoms used by the script.
	weight = uint32(len(w.Script) / AtomSize)
	if len(w.Script)%AtomSize != 0 {
		weight++
	}

	// Add an additional atom for the wallet itself.
	weight++

	// Multiply script and wallet weight by the walletAtomMultiplier to account
	// for the snapshots that the wallet needs to reside in.
	weight *= walletAtomMultiplier

	// Add non-replicated weight according to the size of the wallet sector.
	weight += uint32(w.SectorSettings.Atoms) + w.SectorSettings.UpdateAtoms

	return
}

// walletFilename returns the filename for a wallet, receiving only the id of
// the wallet as input.
func (s *State) walletFilename(id WalletID) (filename string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	filename = s.walletPrefix + "." + suffix
	return
}

// InsertWallet takes a new wallet and inserts it into the wallet tree.
// It returns an error if the wallet already exists within the state.
func (s *State) InsertWallet(w Wallet) (err error) {
	wn := s.walletNode(w.ID)
	if wn != nil {
		err = errors.New("wallet of that id already exists in quorum")
		return
	}

	wn = new(walletNode)
	wn.id = w.ID
	wn.weight = int(w.SectorSettings.Atoms) - int(QuorumSize)
	if wn.weight < 0 {
		wn.weight = 0
	}
	s.insertWalletNode(wn)

	if w.KnownScripts == nil {
		w.KnownScripts = make(map[string]ScriptInputEvent)
	} else {
		for _, scriptEvent := range w.KnownScripts {
			s.InsertEvent(&scriptEvent)
		}
	}

	s.SaveWallet(w)
	return
}

// LoadWallet checks the wallettree for existence of the wallet, and then loads
// the wallet from disk if the wallet exists.
func (s *State) LoadWallet(id WalletID) (w Wallet, err error) {
	// Check that the wallet is in the wallettree.
	wn := s.walletNode(id)
	if wn == nil {
		err = fmt.Errorf("no wallet of id %v exists.", id)
		return
	}

	// Fetch the wallet filename and open the file.
	walletFilename := s.walletFilename(id)
	file, err := os.Open(walletFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

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
		err = fmt.Errorf("no wallet of that id exists: %v", w.ID)
		return
	}
	weightDelta := int(w.SectorSettings.Atoms) - wn.nodeWeight()

	// Ideally, this would never be triggered. Instead, careful resource
	// management in the quorum would prevent a too-heavy wallet from ever
	// getting this far through the insert process.
	if s.walletRoot.weight+weightDelta > AtomsPerQuorum {
		err = errors.New("wallet is too heavy to fit in the quorum")
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

// RemoveWallet removes a Wallet from the wallet tree, and deletes
// the file that contains the wallet.
func (s *State) RemoveWallet(id WalletID) {
	filename := s.walletFilename(id)
	err := os.Remove(filename)
	if err != nil {
		panic(err)
	}
	s.removeWalletNode(id)
}
