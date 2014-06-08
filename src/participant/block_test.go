package participant

import (
	"siacrypto"
	"testing"
)

func TestBlockHistoryHeaderEncoding(t *testing.T) {
	var bhh blockHistoryHeader
	bhh.latestBlock = uint32(siacrypto.RandomUInt64())
	for i := range bhh.blockOffsets {
		bhh.blockOffsets[i] = uint32(siacrypto.RandomUInt64())
	}

	bhhBytes, err := bhh.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	if len(bhhBytes) != BlockHistoryHeaderSize {
		t.Error("Length of bhhBytes is not equal to BlockHistoryHeaderSize")
	}

	var newBhh blockHistoryHeader
	err = newBhh.GobDecode(bhhBytes)
	if err != nil {
		t.Fatal(err)
	}

	if newBhh.latestBlock != bhh.latestBlock {
		t.Error("Encoded and Decoded bhh.latestBlock do not match")
	}
	for i := range bhh.blockOffsets {
		if newBhh.blockOffsets[i] != bhh.blockOffsets[i] {
			t.Error("Encoded and decoded bhh.blockoffsets do not match for index", i)
		}
	}
}

func TestBlockEncoding(t *testing.T) {
}
