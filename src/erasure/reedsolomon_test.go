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
	// acceptable, but data padded to 8*k and not 64*k would cause an error.
	// This checks that the error does not reappear.
	originalData := siacrypto.RandomByteSlice(1080)
	encoded, err := ReedSolomonEncode(9, 3, originalData)
	if err != nil {
		t.Fatal(err)
	}

	// Try to recover the file using only the last 9 pieces, which means all 3
	// non-original pieces will be used.
	remaining := encoded[3:]
	indices := make([]byte, len(remaining))
	for index := range indices {
		indices[index] = byte(index) + 3
	}
	recoveredData, err := ReedSolomonRecover(9, 3, remaining, indices)
	if err != nil {
		t.Fatal(err)
	}

	// Verify 'recoveredData' for consisitency with 'originalData'.
	for i := range originalData {
		if recoveredData[i] != originalData[i] {
			t.Fatal("recovered data does not match original data")
		}
	}
}

// BenchmarkReedSolomonEncode tests the throughput of the ReedSolomonEncode function.
func BenchmarkReedSolomonEncode(b *testing.B) {
	k, m := 51, 128-51
	data := siacrypto.RandomByteSlice(100032 * k) // ~5 MB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReedSolomonEncode(k, m, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReedSolomonEncodeSmall(b *testing.B) {
	k, m := 51, 128-51
	data := siacrypto.RandomByteSlice(64 * 51)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReedSolomonEncode(k, m, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReedSolomonRecover tests the throughput of the ReedSolomonRecover function.
func BenchmarkReedSolomonRecover(b *testing.B) {
	k, m := 51, 128-51
	data := siacrypto.RandomByteSlice(100032 * k)
	encoded, err := ReedSolomonEncode(k, m, data)
	if err != nil {
		b.Fatal(err)
	}
	// use only non-original pieces
	remaining := encoded[m:]
	indices := make([]byte, len(remaining))
	for i := range indices {
		indices[i] = byte(i + m)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = ReedSolomonRecover(k, m, remaining, indices)
		if err != nil {
			b.Fatal(err)
		}
	}
}
