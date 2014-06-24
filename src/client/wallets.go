package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"siaencoding"
)

type Wallet struct {
	ID   uint32
	Type string
}

func SaveWallet(id uint32, walletType string, destFile string) (err error) {
	if destFile == "" {
		fmt.Errorf("Cannot save, file name is empty string")
		return
	}
	var wallet *Wallet
	wallet = new(Wallet)

	wallet.ID = id
	wallet.Type = walletType
	w1 := new(bytes.Buffer)
	encoder := gob.NewEncoder(w1)
	err = encoder.Encode(wallet)
	if err != nil {
		panic(err)
	}
	walletSlice := w1.Bytes()

	walletSize := uint32(len(walletSlice))
	sizeSlice := siaencoding.EncUint32(walletSize)

	f, err := os.Create(destFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(sizeSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(walletSlice)
	if err != nil {
		panic(err)
	}
	return
}

func LoadWallet(fileName string) (wallet *Wallet, err error) {
	if fileName == "" {
		err = fmt.Errorf("Cannot load, file name is empty string")
	}
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	sizeSlice := make([]byte, 4)
	_, err = f.Read(sizeSlice)
	if err != nil {
		panic(err)
	}
	size := siaencoding.DecUint32(sizeSlice)
	walletSlice := make([]byte, size)
	_, err = f.Read(walletSlice)
	if err != nil {
		panic(err)
	}
	wallet = new(Wallet)
	r := bytes.NewBuffer(walletSlice)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&wallet)
	if err != nil {
		panic(err)
	}
	return
}
