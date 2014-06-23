package client

import (
	"fmt"
	"os"
	"siacrypto"
)

func SaveKeyPair(publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, destFile string) (err error) {
	if publicKey == nil || secretKey == nil {
		err = fmt.Errorf("Cannot write nil key to file")
		return
	}
	if destFile == "" {
		err = fmt.Errorf("Cannot save, file name is empty string")
		return
	}
	pubSlice, err := publicKey.GobEncode()
	if err != nil {
		panic(err)
	}
	secSlice, err := secretKey.GobEncode()
	if err != nil {
		panic(err)
	}

	f, err := os.Create(destFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

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

func LoadKeyPair(filePath string) (publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, err error) {
	if filePath == "" {
		err = fmt.Errorf("Cannot load, file name is empty string")
    return
	}
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(pubSlice)
	if err != nil {
		panic(err)
	}
	_, err = f.Read(secSlice)
	if err != nil {
		panic(err)
	}
	publicKey = new(siacrypto.PublicKey)
	secretKey = new(siacrypto.SecretKey)

	err = publicKey.GobDecode(pubSlice)
	if err != nil {
		panic(err)
	}
	err = secretKey.GobDecode(secSlice)
	if err != nil {
		panic(err)
	}
	return
}
