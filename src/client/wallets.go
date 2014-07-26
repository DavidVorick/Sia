package client

import (
	"errors"
	"os"
	"siacrypto"
	"siaencoding"
	"state"
)

func (c *Client) GetGenericWallets() (ids []state.WalletID) {
	ids = make([]state.WalletID, len(c.genericWallets))
	i := 0
	for key, _ := range c.genericWallets {
		ids[i] = key
		i++
	}
	return
}

func (c *Client) EnterWallet(id state.WalletID) (err error) {
	_, exists := c.genericWallets[id]
	if exists {
		c.CurID = id
	} else {
		err = errors.New("Invalid Wallet ID")
	}
	return
}

func SaveWallet(id state.WalletID, keypair *Keypair, destFile string) (err error) {
	if keypair == nil {
		err = errors.New("Cannot encode nil key pair")
		return
	}

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

func LoadWallet(fileName string) (id state.WalletID, keypair *Keypair, err error) {
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
	keypair = new(Keypair)
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
