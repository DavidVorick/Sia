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
	var segments [QuorumSize]io.Writer
	var segmentBuffer [QuorumSize]bytes.Buffer
	for i := range segments {
		segments[i] = &segmentBuffer[i]
	}
	_, err := RSEncode(reader, segments, StandardK)
	if err != nil {
		t.Fatal(err)
	}

	// Create the byte slices continaing the encoded data.
	var segmentBytes [QuorumSize][]byte
	for i := range segmentBytes {
		segmentBytes[i] = segmentBuffer[i].Bytes()
	}

	// Build readers around the encoded data.
	var readers [StandardK]io.Reader
	var indicies []byte
	for i := range readers {
		readers[i] = bytes.NewReader(segmentBytes[i])
		indicies = append(indicies, byte(i))
	}

	// Create an output buffer, and recover the data.
	var output bytes.Buffer
	_, err = RSRecover(readers[:], indicies, &output, StandardK)
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
