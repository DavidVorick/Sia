package script

import (
	"quorum"
	"testing"
)

func TestOpCodes(t *testing.T) {
	// create script execution environment
	q := new(quorum.Quorum)
	q.CreateBootstrapWallet(1, quorum.NewBalance(0, 15000), []byte{0x2D})
	si := &ScriptInput{
		WalletID: 1,
	}

	// test jump
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		// if 2 == 3, jump ahead 10 instructions (causing an error)
		0x16, 0x1F, 0x00, 0x0A,
		0xFF,
	}
	_, err := si.Execute(q)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test store/load
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x04,       // dup 2
		0x21, 0x01, // store 2 in register 1
		0x22, 0x01, // load 2 from register 1
		// if 2 != 2, jump ahead 10 instructions (causing an error)
		0x17, 0x1F, 0x00, 0x0A,
		0xFF,
	}
	_, err = si.Execute(q)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test invalid scripts
	si.Input = []byte{
		0x01, 0xAA,
		0x06, 0xFF,
	}
	_, err = si.Execute(q)
	if err == nil {
		t.Fatal("expected stack empty error")
	}
	si.Input = []byte{
		0x11,
	}
	_, err = si.Execute(q)
	if err == nil {
		t.Fatal("expected missing argument error")
	}
	si.Input = nil
	_, err = si.Execute(q)
	if err == nil {
		t.Fatal("expected missing terminator error")
	}
	si.Input = []byte{
		0x25, 0xFF, 0xF6,
		0xFF,
	}
	_, err = si.Execute(q)
	if err == nil {
		t.Fatal("expected out of bounds error")
	}
}
