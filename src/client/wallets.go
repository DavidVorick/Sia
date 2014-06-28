package client

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siaencoding"
)

func SaveWallet(id quorum.WalletID, keypair *siacrypto.Keypair, destFile string) (err error) {
	if destFile == "" {
		fmt.Errorf("Cannot save, file name is empty string")
		return
	}
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
		panic(err)
	}
	secSlice, err := keypair.SK.GobEncode()
	if err != nil {
		panic(err)
	}

	f, err := os.Create(destFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(idSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(pubSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(secSlice)
	if err != nil {
		panic(err)
	}
	return
}

func LoadWallet(fileName string) (id quorum.WalletID, keypair *siacrypto.Keypair, err error) {
	if fileName == "" {
		err = fmt.Errorf("Cannot load, file name is empty string")
	}
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	idSlice := make([]byte, quorum.WalletIDSize)
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(idSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Read(pubSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Read(secSlice)
	if err != nil {
		panic(err)
	}
	id = quorum.WalletID(siaencoding.DecUint64(idSlice))
	keypair = new(siacrypto.Keypair)
	keypair.PK = new(siacrypto.PublicKey)
	keypair.SK = new(siacrypto.SecretKey)
	err = keypair.PK.GobDecode(pubSlice)
	if err != nil {
		panic(err)
	}
	err = keypair.SK.GobDecode(secSlice)
	if err != nil {
		panic(err)
	}
	return
}
