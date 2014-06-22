package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"siacrypto"
)

func gobEncodeInt(x int) (gobInt []byte, err error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(x)
	if err != nil {
		panic(err)
	}
	gobInt = w.Bytes()
	return
}

func SaveKeyPair(publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, destFile string) (err error) {
	if publicKey == nil || secretKey == nil {
		err = fmt.Errorf("Cannot write nil key to file")
	}
	pubSlice, err := publicKey.GobEncode()
	if err != nil {
		panic(err)
		return
	}
	secSlice, err := secretKey.GobEncode()
	if err != nil {
		panic(err)
		return
	}

	f, err := os.Create(destFile)
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()

	lenPubSlice, err := gobEncodeInt(len(pubSlice))
	_, err = f.Write(lenPubSlice)
	if err != nil {
		panic(err)
		return
	}
	lenSecSlice, err := gobEncodeInt(len(secSlice))
	_, err = f.Write(lenSecSlice)
	if err != nil {
		panic(err)
		return
	}
	_, err = f.Write(pubSlice)
	if err != nil {
		panic(err)
		return
	}
	_, err = f.Write(secSlice)
	if err != nil {
		panic(err)
		return
	}
	return
}
