package client

import (
	"errors"
	"fmt"
	"os"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
)

// Returns a list of all wallets available to the client.
func (c *Client) GetWalletIDs() (ids []state.WalletID) {
	ids = make([]state.WalletID, 0, len(c.genericWallets))
	for id, _ := range c.genericWallets {
		ids = append(ids, id)
	}
	// Add other types of wallets as they are implemented.
	return
}

// Wallet type takes an id as input and returns the wallet type. An error is
// returned if the wallet is not found by the client.
func (c *Client) WalletType(id state.WalletID) (walletType string, err error) {
	// Check if the wallet is a generic type.
	_, exists := c.genericWallets[id]
	if exists {
		walletType = "generic"
		return
	}

	// Check for other types of wallets.

	err = fmt.Errorf("Wallet is not available.")
	return
}

// Iterates through the client and saves all of the wallets to disk.
func (c *Client) SaveAllWallets() (err error) {
	var filename string
	for id, keypair := range c.genericWallets {
		filename = fmt.Sprintf("%x.id", id)
		err = SaveWallet(id, keypair, filename)
		if err != nil {
			return
		}
	}
	// Save other types of wallets as they are implemented.
	return
}

func SaveWallet(id state.WalletID, keypair Keypair, destFile string) (err error) {
	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(siaencoding.EncUint64(uint64(id)))
	if err != nil {
		return
	}
	_, err = f.Write(keypair.PK[:])
	if err != nil {
		return
	}
	_, err = f.Write(keypair.SK[:])
	if err != nil {
		return
	}
	return
}

func LoadWallet(fileName string) (id state.WalletID, keypair Keypair, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	idSlice := make([]byte, 8)
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(idSlice)
	if err != nil {
		return
	}
	_, err = f.Read(pubSlice)
	if err != nil {
		return
	}
	_, err = f.Read(secSlice)
	if err != nil {
		return
	}
	id = state.WalletID(siaencoding.DecUint64(idSlice))
	if copy(keypair.PK[:], pubSlice) != siacrypto.PublicKeySize {
		err = errors.New("bad public key length")
		return
	}
	if copy(keypair.SK[:], secSlice) != siacrypto.SecretKeySize {
		err = errors.New("bad secret key length")
		return
	}

	return
}
