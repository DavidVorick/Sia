// Sia uses Reed-Solomon coding for error correction. This package has no
// method for error detection however, so error detection must be performed
// elsewhere.
//
// We use the repository 'NebulousLabs/longhair' to handle the erasure coding.
// As far as I'm aware, it's the fastest library that's open source.
// It is a fork of 'catid/longhair', and we intend to merge all changes from
// the original.
//
// Longhair is a c++ library. Here, it is cast to a C library and then called
// using cgo.
package erasure

// #include "erasure.c"
import "C"

import (
	"fmt"
	"unsafe"
)

// EncodeRedundancy takes a []byte and some redundancy parameters and produces
// a set of k + m byte slices that compose the encoded data
func EncodeRedundancy(k byte, m byte, decoded []byte) (encoded [][]byte, err error) {
	// check for nil values
	if decoded == nil {
		err = fmt.Errorf("EncodeRedundancy: received nil input!")
		return
	}

	// check that k and m are resonable values
	if k < 1 || m < 0 {
		err = fmt.Errorf("EncodeRedundancy: minimum value for k is 1 and for m is 0")
		return
	}
	if k+m > 255 {
		err = fmt.Errorf("EncodeRedundancy: k + m must be less than 255")
		return
	}

	// check for correct padding on decoded []byte
	if len(decoded)%int((k*8)) != 0 {
		err = fmt.Errorf("EncodeRedundancy: input has not been properly padded!")
		return
	}

	b := len(decoded) / int(k)

	// call longhair to do encoding
	redundantChunk := C.encodeRedundancy(C.int(k), C.int(m), C.int(b), (*C.char)(unsafe.Pointer(&decoded[0])))
	redundantBytes := C.GoBytes(unsafe.Pointer(redundantChunk), C.int(int(m)*b))

	// allocate encoded
	encoded = make([][]byte, k+m)

	// split the original data into encoded
	for i := 0; i < int(k); i++ {
		encoded[i] = decoded[i*b : (i+1)*b]
	}
	for i := 0; i < int(m); i++ {
		encoded[i+int(k)] = redundantBytes[i*b : (i+1)*b]
	}

	// free the memory allocated by the C call
	C.free(unsafe.Pointer(redundantChunk))

	return
}

// Recover takes a set of pieces at least k in length, along with a list of
// which index each piece was originally in the encoded data, and uses these
// variables to produce a []byte that is equivalent to the original data.
func Recover(k int, m int, remaining [][]byte, indicies []int) (recovered []byte, err error) {
	// check for nil values
	if remaining == nil || indicies == nil {
		err = fmt.Errorf("Recover: received nil input")
		return
	}

	// check that k and m are reasonable values
	if k < 1 || m < 0 {
		err = fmt.Errorf("Recover: minimum value for k is 1 and m is 0")
		return
	}
	if k+m > 255 {
		err = fmt.Errorf("Recover: k + m must be less than 255")
		return
	}
	if len(remaining) < k {
		err = fmt.Errorf("Recover: insufficient pieces to recover original")
		return
	}

	// check for reasonable values within remaining
	if remaining[0] == nil {
		err = fmt.Errorf("Recover: received nil slice within set of data")
		return
	}
	b := len(remaining[0])
	if b == 0 {
		err = fmt.Errorf("Recover: cannot recover empty data")
		return
	}
	if b%8 != 0 {
		err = fmt.Errorf("Recover: received input that is not modulo 8")
		return
	}
	for i := 0; i < k; i++ {
		if remaining[i] == nil {
			err = fmt.Errorf("Recover: received nil slice within set of data")
			return
		}
		if len(remaining[i]) != b {
			err = fmt.Errorf("Recover: chunk sizes are not consistent within input")
			return
		}
	}

	// check for reasonable and unique values within indicies
	if len(indicies) < k {
		err = fmt.Errorf("Recover: Indicies does not contain enough indexes")
		return
	}
	seenIndicies := make(map[int]bool)
	for i := 0; i < k; i++ {
		if indicies[i] >= k+m || indicies[i] < 0 {
			err = fmt.Errorf("Recover: Received an index that is out of bounds")
			return
		}
		seen := seenIndicies[indicies[i]]
		if seen {
			err = fmt.Errorf("Recover: repeat indicies presented")
			return
		}
		seenIndicies[indicies[i]] = true
	}

	// copy all data into a single slice
	recovered = make([]byte, k*b)
	remainingIndicies := make([]uint8, k)
	for i := 0; i < k; i++ {
		copy(recovered[i*b:(i+1)*b], remaining[i])
		remainingIndicies[i] = uint8(indicies[i])
	}

	C.recover(C.int(k), C.int(m), C.int(b), (*C.uchar)(unsafe.Pointer(&recovered[0])), (*C.uchar)(unsafe.Pointer(&remainingIndicies[0])))

	return
}
