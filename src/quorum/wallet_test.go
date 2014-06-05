package quorum

import (
	"siacrypto"
	"testing"
)

func TestIdEncoding(t *testing.T) {
	randomBytes, err := siacrypto.RandomByteSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	var w0 walletHandle
	copy(w0[:], randomBytes)
	w0ID := w0.id()
	w0Handle := w0ID.handle()
	w0Confirm := w0Handle.id()

	if w0 != w0Handle {
		t.Error("Encoding Mismatch:", w0, ":", w0Handle)
	}
	if w0ID != w0Confirm {
		t.Error("Encoding Mismatch:", w0ID, ":", w0Confirm)
	}
}
