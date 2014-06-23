package client

import (
	"fmt"
	"os"
	"siacrypto"
)

func SaveKeyPair(publicKey *siacrypto.PublicKey, secretKey *siacrypto.SecretKey, destFile string) (err error) {
	if publicKey == nil || secretKey == nil {
		err = fmt.Errorf("Cannot write nil key to file")
	}
	if destFile == "" {
		err = fmt.Errorf("Cannot save, file name is empty string")
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
		err = fmt.Errorf("Cannot load, file name is empty string")
	}
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
		return
	}
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(pubSlice)
	if err != nil {
		panic(err)
		return
	}
  _, err = f.Read(secSlice)
	if err != nil {
		panic(err)
		return
	}
	publicKey = new(siacrypto.PublicKey)
	secretKey = new(siacrypto.SecretKey)

	err = publicKey.GobDecode(pubSlice)
	if err != nil {
		panic(err)
		return
	}
	err = secretKey.GobDecode(secSlice)
	if err != nil {
		panic(err)
		return
	}
	return
}
