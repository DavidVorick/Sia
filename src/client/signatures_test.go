package client

import (
	"os"
	"siacrypto"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	pk0, sk0, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	err = SaveKeyPair(pk0, sk0, "tempFile")
	pk1, sk1, err := LoadKeyPair("tempFile")

	if pk0.Compare(pk1) == false || sk0.Compare(sk1) == false {
		t.Error("Keys not sved or loaded correctly")
	}
	err = os.Remove("tempFile")
}
