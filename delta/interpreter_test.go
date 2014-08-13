package delta

import (
	"testing"

	//"siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

// initialize a script execution environment
func initEnv() (e Engine, si ScriptInput) {
	e.state.SetWalletPrefix(siafiles.TempFilename("InterpreterTest"))
	// create a wallet that immediately passes control to its input
	e.state.InsertWallet(state.Wallet{
		ID:      1,
		Balance: state.NewBalance(0, 15000),
		Script:  []byte{0x38},
	})
	si = ScriptInput{
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
	_, err := e.Execute(si)
	if err != nil {
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
	_, err = e.Execute(si)
	if err != nil {
		t.Error(err)
	}
}

// check that invalid scripts produce an error
func TestInvalidScripts(t *testing.T) {
	e, si := initEnv()

	si.Input = []byte{
		0x01, 0xAA, 0x06,
	}
	_, err := e.Execute(si)
	if err == nil {
		t.Error("expected stack empty error")
	}
	si.Input = []byte{
		0x11,
	}
	_, err = e.Execute(si)
	if err == nil {
		t.Error("expected missing argument error")
	}
	si.Input = []byte{
		0x25, 0xFF, 0x7F,
	}
	_, err = e.Execute(si)
	if err == nil {
		t.Error("expected out of bounds error")
	}
}

// test the "verify" opcode
/*
func TestVerify(t *testing.T) {
	e, si := initEnv()

	// generate public key
	publicKey, secretKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Error(err)
	}
	gobPKey, err := publicKey.GobEncode()
	if err != nil {
		t.Error(err)
	}

	// generate signed message
	message := []byte("test")
	signedMessage, err := secretKey.Sign(message)
	if err != nil {
		t.Error(err)
	}
	encSm, err := signedMessage.GobEncode()
	if err != nil {
		t.Error(err)
	}

	// construct script
	si.Input = []byte{
		0x25, 0x0D, 0x00, // move data pointer to start of public key
		0x39, 0x20, 0x01, // copy public key into buffer 1
		0x2E, 0x02, //       copy signed message into buffer 2
		0x34, 0x01, 0x02, // verify signature
		0x38, //             if invalid signature, reject
		0xFF, //             otherwise, exit normally
	}
	si.Input = append(si.Input, gobPKey...)
	si.Input = append(si.Input, encSm...)

	// execute script
	_, err = e.Execute(si)
	if err != nil {
		t.Error(err)
	}
}
*/
