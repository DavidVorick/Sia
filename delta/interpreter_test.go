package delta

import (
	"testing"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/sialog"
	"github.com/NebulousLabs/Sia/state"
)

// initialize a script execution environment
func initEnv() (e Engine, si state.ScriptInput) {
	e.SetLogger(sialog.Default)
	e.state.SetWalletPrefix(siafiles.TempFilename("InterpreterTest"))
	// create a wallet that immediately passes control to its input
	e.state.InsertWallet(state.Wallet{
		ID:      1,
		Balance: state.NewBalance(15000),
		Script:  []byte{0x38}, // transfer control to input
	}, true)
	si = state.ScriptInput{
		WalletID: 1,
	}
	return
}

// test basic stack operations, jumps, and store/load
func TestOpCodes(t *testing.T) {
	e, si := initEnv()

	// test jump
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		// if 2 == 3, jump ahead 10 instructions (causing an error)
		0x16, 0x1F, 0x0A, 0x00,
	}
	if err := e.Execute(si); err != nil {
		t.Error(err)
	}

	// test store/load
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x04,       // dup 2
		0x30, 0x01, // store 2 in register 1
		0x31, 0x01, // load 2 from register 1
		// if 2 != 2, jump ahead 10 instructions (causing an error)
		0x17, 0x1F, 0x0A, 0x00,
	}
	if err := e.Execute(si); err != nil {
		t.Error(err)
	}

	// test data_seek
	si.Input = []byte{
		0xE6, 0xFE, // seek past next 0xFE
		0xE6, 0xFE, // seek past next 0xFE
		0xE6, 0xFE, // seek past next 0xFE
		0x38, //       transfer
		0xFE, //       reject (error)
		0xFE, //       reject (error)
		0xFE, //       reject (error)
		0xFF, //       terminate (no error)
	}
	if err := e.Execute(si); err != nil {
		t.Error(err)
	}
}

// check that invalid scripts produce an error
func TestInvalidScripts(t *testing.T) {
	e, si := initEnv()

	si.Input = []byte{
		0x01, 0xAA, 0x06,
	}
	if e.Execute(si) == nil {
		t.Error("expected stack empty error")
	}
	si.Input = []byte{
		0x11,
	}
	if e.Execute(si) == nil {
		t.Error("expected missing argument error")
	}
	si.Input = []byte{
		0x21, 0xFF, 0x7F,
	}
	if e.Execute(si) == nil {
		t.Error("expected out of bounds error")
	}
}

// test the "verify" opcode
func TestVerify(t *testing.T) {
	e, si := initEnv()

	// generate public key
	publicKey, secretKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// generate signature
	message := []byte("test")
	signature, err := secretKey.Sign(message)
	if err != nil {
		t.Fatal(err)
	}

	// construct script
	si.Input = appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to start of public key
			0x34, 0x20, // push public key
			0x34, 0x40, // push signature
			0xE4, //       push message
			0x40, //       verify signature
			0xE5, //       if invalid signature, reject
			0xFF, //       otherwise, exit normally
		},
		publicKey[:],
		signature[:],
		message,
	)

	if err := e.Execute(si); err != nil {
		t.Error(err)
	}
}

// exhaust various resources
func TestExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	e, si := initEnv()

	// exhaust balance by executing many expensive instructions
	si.Input = []byte{
		0x21, 0x01, 0x00, // goto 01 (this instruction; i.e. an infinite loop)
		0xFF, //             unreachable (hopefully...)
	}

	if e.Execute(si) == nil {
		t.Error("expected resource exhaustion error")
	}

	// currently there is no way to exhaust instBalance; instBalance and
	// costBalance are both set to 10000, and there are no 0-cost instructions
	// (besides op_exit and op_reject)

	// exceed memory usage limit by filling up registers
	si.Input = []byte{
		0x02, 0xFF, 0xFF, // push 2^16
		0x36, 0x01, //       copy 2^16 bytes into register 1
		0xFF, //             unreachable (hopefully...)
	}

	if e.Execute(si) == nil {
		t.Error("expected resource exhaustion error")
	}
}
