package participant

import (
	"quorum"
	"quorum/script"
	"siacrypto"
	"testing"
)

// Create a block history header with random values, then encode and decode it
// and check for consistency
func TestBlockHistoryHeaderEncoding(t *testing.T) {
	// create blockHistoryHeader and check for random values
	var bhh blockHistoryHeader
	bhh.latestBlock = uint32(siacrypto.RandomUInt64())
	for i := range bhh.blockOffsets {
		bhh.blockOffsets[i] = uint32(siacrypto.RandomUInt64())
	}

	// encode it
	bhhBytes, err := bhh.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// verify consistency of constant
	if len(bhhBytes) != BlockHistoryHeaderSize {
		t.Error("Length of bhhBytes is not equal to BlockHistoryHeaderSize")
	}

	// decode it
	var newBhh blockHistoryHeader
	err = newBhh.GobDecode(bhhBytes)
	if err != nil {
		t.Fatal(err)
	}

	// check values for consistency
	if newBhh.latestBlock != bhh.latestBlock {
		t.Error("Encoded and Decoded bhh.latestBlock do not match")
	}
	for i := range bhh.blockOffsets {
		if newBhh.blockOffsets[i] != bhh.blockOffsets[i] {
			t.Error("Encoded and decoded bhh.blockoffsets do not match for index", i)
		}
	}
}

// Create a block with random values, then encode it and decode it and check
// the decoded block against the encoded block for consistency
func TestBlockEncoding(t *testing.T) {
	// create random heartbeats to populate the block with
	var hb0 heartbeat
	var hb1 heartbeat
	copy(hb0.entropy[:], siacrypto.RandomByteSlice(quorum.EntropyVolume))
	copy(hb1.entropy[:], siacrypto.RandomByteSlice(quorum.EntropyVolume))
	hb0.scriptInputs = make([]script.ScriptInput, 1)
	hb0.scriptInputs[0] = script.ScriptInput{
		WalletID: quorum.WalletID(siacrypto.RandomUInt64()),
		Input:    siacrypto.RandomByteSlice(20),
	}
	hb1.scriptInputs = make([]script.ScriptInput, 1)
	hb1.scriptInputs[0] = script.ScriptInput{
		WalletID: quorum.WalletID(siacrypto.RandomUInt64()),
		Input:    siacrypto.RandomByteSlice(18),
	}

	// create the block and fill it with random values
	var b block
	b.height = uint32(siacrypto.RandomUInt64())
	copy(b.parent[:], siacrypto.RandomByteSlice(siacrypto.TruncatedHashSize))
	b.heartbeats[1] = &hb0
	b.heartbeats[3] = &hb1

	// encode the block
	blockBytes, err := b.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// decode the block
	var db block
	err = db.GobDecode(blockBytes)
	if err != nil {
		t.Fatal(err)
	}

	// compare height and parent
	if db.height != b.height {
		t.Error("height not encoded and decoded as equivalent")
	}
	if db.parent != b.parent {
		t.Error("parent not encoded and decoded as equivalent")
	}

	// check that the correct heartbeats are nil
	if db.heartbeats[0] != nil || db.heartbeats[2] != nil {
		t.Error("heartbeats are being loaded into the incorrect slots")
	}

	// check that the entropies match
	if db.heartbeats[1].entropy != b.heartbeats[1].entropy {
		t.Error("entropy is being encoded or decoded incorrectly")
	}
	if db.heartbeats[3].entropy != b.heartbeats[3].entropy {
		t.Error("entropy is being encoded or decoded incorrectly")
	}
}
