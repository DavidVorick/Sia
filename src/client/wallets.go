package client

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siaencoding"
)

func SaveWallet(id quorum.WalletID, keypair *siacrypto.Keypair, destFile string) (err error) {
	if keypair == nil {
		fmt.Errorf("Cannot encode nil key pair")
		return
	}
	if keypair.PK == nil || keypair.SK == nil {
		fmt.Errorf("Cannot write nil key to file")
		return
	}

	idSlice := siaencoding.EncUint64(uint64(id))
	pubSlice, err := keypair.PK.GobEncode()
	if err != nil {
		return
	}
	secSlice, err := keypair.SK.GobEncode()
	if err != nil {
		return
	}

	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(idSlice)
	if err != nil {
		return
	}
	_, err = f.Write(pubSlice)
	if err != nil {
		return
	}
	_, err = f.Write(secSlice)
	if err != nil {
		return
	}
	return
}

func LoadWallet(fileName string) (id quorum.WalletID, keypair *siacrypto.Keypair, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	idSlice := make([]byte, quorum.WalletIDSize)
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
	id = quorum.WalletID(siaencoding.DecUint64(idSlice))
	keypair = new(siacrypto.Keypair)
	keypair.PK = new(siacrypto.PublicKey)
	keypair.SK = new(siacrypto.SecretKey)
	err = keypair.PK.GobDecode(pubSlice)
	if err != nil {
		return
	}
	err = keypair.SK.GobDecode(secSlice)
	if err != nil {
		return
	}
	return
}
