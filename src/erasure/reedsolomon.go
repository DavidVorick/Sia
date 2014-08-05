package erasure

// The default encoding for files upload to Sia is Reed-Solomon coding. We use
// the repository 'NebulousLabs/longhair' as a library for the Reed-Solomong
// coding. The repository is a fork of 'catid/longhair', and we intend to keep
// our repo up-to-date with the upstream repo.
//
// Because longhair has not been audited, and also because we aren't entirely
// sure of our approach to file recovery, Reed-Solomon coding will not be run
// by the server, instead only by the client.
//
// reedsolomon.go does not provide any tools for error detection, only tools
// for error correction. Error detection must be performed by another set of
// code.
//
// reedsolomon.go and reedsolomon.c should be viewed as one file. reedsolomon.c
// should never be accessed or used by a file other than reedsolomon.go.

// #include "reedsolomon.c"
import "C"

import (
	"errors"
	"unsafe"
)

// ReedSolomonEncode takes as input 'k', 'm', and a []byte 'original' that
// represents the original data. 'k' is the number of encoded pieces required
// to recover the original data. 'm' is the number of redundant pieces. 'k + m'
// is the total number of pieces. 'k' must be greater than 1, 'm' must be
// greater than 0, and their sum must be less than 256. 'original' must have a
// length that is divisible by 'k * 8'.
//
// Returned is a set of 'k + m' byte slices containing the encoded data. Each
// will be of length (len(original) / k). Any 'k' of the encoded pieces can be
// used to recover 'original'. If there is an error, 'encoded' should be
// discarded.
func ReedSolomonEncode(k, m int, original []byte) (encoded [][]byte, err error) {
	// Check for nil and zero values within the input.
	if original == nil {
		err = errors.New("received nil input")
		return
	}
	if len(original) == 0 {
		err = errors.New("cannot encode a slice of length 0")
		return
	}

	// Check that k and m fit the function requirements.
	if k < 1 {
		err = errors.New("k must be at least 1")
		return
	}
	if m < 0 {
		err = errors.New("m must be at greater than 0")
		return
	}
	if k+m >= 256 {
		err = errors.New("k + m must be less than 256")
		return
	}

	// Check that 'original' has been correctly padded
	if len(original)%(k*8) != 0 {
		err = errors.New("input has not been properly padded")
		return
	}

	b := len(original) / k

	// Call longhair to perform encoding.
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(b), (*C.char)(unsafe.Pointer(&original[0])))
	if redundantChunk == nil {
		err = errors.New("call to cauchy_256_encode failed")
		return
	}
	redundantBytes := C.GoBytes(unsafe.Pointer(redundantChunk), C.int(m*b))

	// Allocate 'encoded'.
	encoded = make([][]byte, k+m)

	// Slice the data returned from longhair into 'encoded'.
	for i := 0; i < k; i++ {
		encoded[i] = original[i*b : (i+1)*b]
	}
	for i := 0; i < m; i++ {
		encoded[i+k] = redundantBytes[i*b : (i+1)*b]
	}

	// Free the memory allocated by longhair.
	C.free(unsafe.Pointer(redundantChunk))

	return
}

// ReedSolomonRecover takes as input 'k' and 'm', which need to match the 'k'
// and 'm' used during 'ReedSolomonEncode'. 'remaining' is a set of 'k' pieces
// that are being used to recover the data used as 'Original' in
// ReedSolomonEncode. 'indices' maps the relationship between the pieces
// provided in 'remaining' to their original index in 'encoded' from
// ReedSolomonEncode.
//
// 'recovered' should be identical to the input 'original' from
// ReedSolomonEncode. If an error is returned, 'recovered' should be discarded.
func ReedSolomonRecover(k, m int, remaining [][]byte, indices []byte) (recovered []byte, err error) {
	// Check for nil values. The length of 'remaining' and 'indices' are
	// checked after 'k' and 'm' are checked.
	if remaining == nil || indices == nil {
		err = errors.New("received nil input")
		return
	}

	// Check that k and m fit the function requirements.
	if k < 1 {
		err = errors.New("k must be at least 1")
		return
	}
	if m < 0 {
		err = errors.New("m must be at greater than 0")
		return
	}
	if k+m >= 256 {
		err = errors.New("k + m must be less than 256")
		return
	}

	// Check that 'remaining' and 'indices' contain at least 'k' elements.
	if len(remaining) < k {
		err = errors.New("insufficient pieces to recover original")
		return
	}
	if len(indices) < k {
		err = errors.New("insufficient indices")
		return
	}

	// Check that 'remaining' has legal values and will not panic the program.
	if remaining[0] == nil {
		err = errors.New("received nil slice within set of data")
		return
	}
	b := len(remaining[0])
	if b == 0 {
		err = errors.New("cannot recover empty data")
		return
	}
	if b%8 != 0 {
		err = errors.New("remaining pieces do not match padding, should be padded to 8 bytes")
		return
	}
	for i := 0; i < k; i++ {
		if remaining[i] == nil {
			err = errors.New("received nil slice within set of data")
			return
		}
		if len(remaining[i]) != b {
			err = errors.New("sizes of remaining pieces are not consistent")
			return
		}
	}

	// Check that indices has a set of unique values.
	seenIndices := make(map[byte]bool)
	for i := 0; i < k; i++ {
		if int(indices[i]) >= k+m {
			err = errors.New("received out of bounds index")
			return
		}
		if seenIndices[indices[i]] {
			err = errors.New("received duplicate index")
			return
		}
		seenIndices[indices[i]] = true
	}

	// Arrange the data so that longhair will order the data into the single
	// slice 'recovered'
	recovered = make([]byte, k*b)
	remainingIndices := make([]byte, k)
	for i := 0; i < int(k); i++ {
		copy(recovered[i*b:(i+1)*b], remaining[i])
		remainingIndices[i] = indices[i]
	}

	errCode := C.recover(C.int(k), C.int(m), C.int(b), (*C.uchar)(&recovered[0]), (*C.uchar)(&remainingIndices[0]))

	if errCode != 0 {
		err = errors.New("call to cauchy_256_decode failed")
	}

	return
}
