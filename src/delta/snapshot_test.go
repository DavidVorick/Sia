package delta

import (
	"reflect"
	"siacrypto"
	"state"
	"testing"
)

// TestSnapshotOffsetTableEncoding creates, encodes, and then decodes a
// snapshotOffsetTable, and tests for equivalence.
func TestSnapshotOffsetTableEncoding(t *testing.T) {
	// Create a snapshot offset table with different values for every field.
	sot := snapshotOffsetTable{
		stateMetadataOffset:     1,
		stateMetadataLength:     2,
		walletLookupTableOffset: 3,
		walletLookupTableLength: 4,
		eventLookupTableOffset:  5,
		eventLookupTableLength:  6,
	}

	// Encode and decode the snapshot offset table.
	encodedSOT, err := sot.encode()
	if err != nil {
		t.Fatal(err)
	}
	var dsot snapshotOffsetTable
	err = dsot.decode(encodedSOT)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the decoded table to the original table.
	if dsot != sot {
		t.Error("Encoded and decoded snapshotOffsetTable does not equal the original.")
	}
}

// TestWalletOffsetEncoding creates, encodes, and decodes a walletOffset,
// comparing the decoded walletOffset to the original in a test for
// equivalence.
func TestWalletOffsetEncoding(t *testing.T) {
	// Create a walletOffset with different values for every field.
	wo := walletOffset{
		id:     1,
		offset: 2,
		length: 3,
	}

	// Encode and then decode the wallet offset.
	encodedWO, err := wo.encode()
	if err != nil {
		t.Fatal(err)
	}
	var dwo walletOffset
	err = dwo.decode(encodedWO)
	if err != nil {
		t.Fatal(err)
	}

	// Compare decoded wallet offset to the original.
	if dwo != wo {
		t.Error("Encoded and decoded walletOffset does not match the original.")
	}
}

// TestSnapshotProcess creates an engine and fills out all the major fields
// that get saved by a snapshot. Then SaveSnapshot() is called and each piece
// is loaded individually to test for proper retrieveability.
func TestSnapshotProcess(t *testing.T) {
	// Crudely create an engine with all metadata variables filled out.
	e := Engine{
		state: state.State{
			Metadata: state.StateMetadata{
				Germ:         state.Entropy{1},
				Seed:         state.Entropy{2},
				EventCounter: 3,
				StoragePrice: state.NewBalance(4, 5),
				ParentBlock:  siacrypto.Hash{6},
				Height:       7,
			},
		},
		recentHistoryHead: ^uint32(0),
	}
	e.Initialize("../../filesCreatedDuringTesting/TestSnapshotProcess")

	// Save a handful of wallets into the quorum.
	w1 := state.Wallet{
		ID:      8,
		Balance: state.NewBalance(9, 10),
		SectorSettings: state.SectorSettings{
			Atoms: 11,
			K:     12,
			Hash:  siacrypto.Hash{13},
		},
		Script: []byte{14, 15, 16},
	}
	e.InsertWallet(w1)

	w2 := state.Wallet{
		ID:      17,
		Balance: state.NewBalance(18, 19),
		SectorSettings: state.SectorSettings{
			Atoms: 20,
			K:     21,
			Hash:  siacrypto.Hash{22},
		},
		Script: []byte{23, 24, 25},
	}
	e.InsertWallet(w2)

	// Add some extra wallets to confirm that the binary search works.
	for i := 0; i < 32; i++ {
		w := w2
		w.ID = state.WalletID(i+50)
		e.InsertWallet(w)
	}

	// Save the snapshot.
	err := e.saveSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Load each component of the snapshot and see if it matches the original.
	metadata, err := e.LoadSnapshotMetadata(0)
	if err != nil {
		t.Fatal(err)
	}
	if metadata != e.state.Metadata {
		t.Error("Upon loading from snapshot, metadata does not equal original metadata")
	}

	// Load the list of wallets from the snapshot and verify for accuracy.
	walletList, err := e.LoadSnapshotWalletList(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(walletList) != 34 {
		t.Error("Expecting 32 wallets in snapshot, got ", len(walletList))
	}
	if walletList[0] != 8 {
		t.Error("Expecting first wallet to have an id of 8, got", walletList[0])
	}
	if walletList[1] != 17 {
		t.Error("Expecting second wallet to have an id of 17, got", walletList[1])
	}

	// Load each wallet individually and check for reachability.
	for _, id := range walletList {
		_, err := e.LoadSnapshotWallet(0, id)
		if err != nil {
			t.Fatal(err)
		}
	}

	// For wallets 1 and 2 in particular, do a deep equals check against the
	// original.
	wallet1, err := e.LoadSnapshotWallet(0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wallet1, w1) {
		t.Error("Wallet does not match its pre-snapshot counterpart.")
	}

	wallet2, err := e.LoadSnapshotWallet(0, 17)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wallet2, w2) {
		t.Error("Wallet does not match its pre-snapshot counterpart.")
	}

	// Events will be implemented later.
}
