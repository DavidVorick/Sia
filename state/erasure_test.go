package state

import (
	"bytes"
	"io"
	"testing"

	"github.com/NebulousLabs/Sia/siacrypto"
)

func TestRSEncodeAndRSDecode(t *testing.T) {
	// Create data to use as input for the encoder.
	input := siacrypto.RandomByteSlice(78)
	reader := bytes.NewReader(input)

	// Create segments for the output.
	segments := make([][]byte, QuorumSize)
	_, err := RSEncode(reader, segments, StandardK)
	if err != nil {
		t.Fatal(err)
	}

	readers := make([]io.Reader, StandardK)
	var indices []byte
	for i := range readers {
		readers[i] = bytes.NewReader(segments[i])
		indices = append(indices, byte(i))
	}

	// Create an output buffer, and recover the data.
	var output bytes.Buffer
	_, err = RSRecover(readers, indices, &output, StandardK)
	if err != nil {
		t.Fatal(err)
	}

	var i int
	outputBytes := output.Bytes()
	for i = 0; i < len(input); i++ {
		if input[i] != outputBytes[i] {
			t.Error("mismatch on byte ", i)
		}
	}

	for i < len(outputBytes) {
		if outputBytes[i] != 0 {
			t.Error("non-empty output found at tail.")
		}
		i++
	}
}
