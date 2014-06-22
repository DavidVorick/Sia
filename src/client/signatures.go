package client

import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
)

func SaveKeyPair(publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, destFile string) (err error) {
	if publicKey == nil || secretKey == nil || destFile == "" {
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

	lenPubSlice := siaencoding.EncUint32(uint32(len(pubSlice)))
	_, err = f.Write(lenPubSlice)
	if err != nil {
		panic(err)
		return
	}
	lenSecSlice := siaencoding.EncUint32(uint32(len(secSlice)))
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

func LoadKeyPair(filePath string) (publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, err error) {
	if filePath == "" {
		err = fmt.Errorf("Cannot load from nil file")
	}
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
		return
	}
	byteLenPub := make([]byte, 4)
	byteLenSec := make([]byte, 4)
	_, err = f.Read(byteLenPub)
	if err != nil {
		panic(err)
		return
	}
	_, err = f.Read(byteLenSec)
	if err != nil {
		panic(err)
		return
	}
	bytePubSlice := make([]byte, siaencoding.DecUint32(byteLenPub))
	byteSecSlice := make([]byte, siaencoding.DecUint32(byteLenSec))
	_, err = f.Read(bytePubSlice)
	_, err = f.Read(byteSecSlice)

	publicKey = new(siacrypto.PublicKey)
	secretKey = new(siacrypto.SecretKey)

	err = publicKey.GobDecode(bytePubSlice)
	fmt.Println("lol")
	if err != nil {
		panic(err)
		return
	}
	err = secretKey.GobDecode(byteSecSlice)
	fmt.Println("lollll")
	err = nil
	if err != nil {
		panic(err)
		return
	}
	return
}
