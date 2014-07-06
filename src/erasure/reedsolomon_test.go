package erasure

import (
	"siacrypto"
	"testing"
)

// TestReedSolomonEncode creates a random byte slice, encodes a set of 12
// pieces with 'k' = 9, and then decodes the file using the 3 tail pieces. This
// test function only checks the general case, and doesn't probe for bugs or
// errors.
func TestReedSolomonEncodeAndRecover(t *testing.T) {
	// Make a random byte slice of length 1080 and encode it into 12 pieces, k=9
	// and m=3. 1080 has been chosen because it is divisible by 8*12 but not
	// 64*12. There used to be a bug where all data padded to 64*k was
	// acceptibile, but data padded to 8*k and not 64*k would cause an error.
	// This checks that the error does not reappear.
	originalData := siacrypto.RandomByteSlice(1080)
	encoded, err := ReedSolomonEncode(9, 3, originalData)
	if err != nil {
		t.Fatal(err)
	}

	// Try to recover the file using only the last 9 pieces, which means all 3
	// non-original pieces will be used.
	remaining := encoded[3:]
	indicies := make([]byte, len(remaining))
	for index := range indicies {
		indicies[index] = byte(index) + 3
	}
	recoveredData, err := ReedSolomonRecover(9, 3, remaining, indicies)

	// Verify 'recoveredData' for consisitency with 'originalData'.
	for i := range recoveredData {
		if recoveredData[i] != originalData[i] {
			t.Error("recovered data does not match original data")
		}
	}
}
