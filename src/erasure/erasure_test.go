package erasure

import (
	"siacrypto"
	"testing"
)

func TestEncodeAndRecover(t *testing.T) {
	originalData := siacrypto.RandomByteSlice(960)
	decoded, err := EncodeRedundancy(8, 3, originalData)
	if err != nil {
		t.Fatal(err)
	}

	corrupted := decoded[3:]
	indicies := make([]byte, len(corrupted))
	for index := range indicies {
		indicies[index] = index+3
	}

	recoveredData, err := Recover(8, 3, corrupted, indicies)

	for i := range recoveredData {
		if recoveredData[i] != originalData[i] {
			t.Fatal("recovered data does not match original data")
		}
	}
}
