package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"siaencoding"
)

type Wallet struct {
	ID         int
	ScriptTags map[string]struct{}
}

func SaveWallet(id int, tags []string, destFile string) (err error) {
	if destFile == "" {
		fmt.Errorf("Cannot save, file name is empty string")
		return
	}
	var wallet *Wallet
	wallet = new(Wallet)

	wallet.ID = id
	for t := range tags {
		wallet.ScriptTags[tags[t]] = struct{}{}
	}
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

func LoadWallet(fileName string) {

}
