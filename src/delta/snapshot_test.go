package delta

import (
	"testing"
)

// TestSnapshotOffsetTableEncoding creates, encodes, and then decodes a
// snapshotOffsetTable, and tests for equivalence.
func TestSnapshotOffsetTableEncoding(t *testing.T) {
	// Create a snapshot offset table with different values for every field.
	sot := snapshotOffsetTable{
		stateMetadataOffset: 1,
		stateMetadataLength: 2,
		walletLookupTableOffset: 3,
		walletLookupTableLength: 4,
		eventLookupTableOffset: 5,
		eventLookupTableLength: 6,
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
		id: 1,
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
