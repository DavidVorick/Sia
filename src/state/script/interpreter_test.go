package script

import (
	"os"
	"siacrypto"
	"state"
	"testing"
)

// initialize a script execution environment
func initEnv() (q *quorum.Quorum, si *ScriptInput) {
	// create script execution environment
	q = new(quorum.Quorum)
	wd, _ := os.Getwd()
	wd += "/../../../participantStorage/InterpreterTest."
	q.SetWalletPrefix(wd)
	q.CreateBootstrapWallet(1, quorum.NewBalance(0, 15000), []byte{0x2F})
	si = &ScriptInput{
		WalletID: 1,
	}
	return
}

// test basic stack operations, jumps, and store/load
func TestOpCodes(t *testing.T) {
	q, si := initEnv()

	// test jump
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		// if 2 == 3, jump ahead 10 instructions (causing an error)
		0x16, 0x1F, 0x0A, 0x00,
	}
	_, err := si.Execute(q)
	if err != nil {
		t.Fatal(err)
	}

	// test store/load
	si.Input = []byte{
		0x01, 0x02, // push 2
		0x04,       // dup 2
		0x21, 0x01, // store 2 in register 1
		0x22, 0x01, // load 2 from register 1
		// if 2 != 2, jump ahead 10 instructions (causing an error)
		0x17, 0x1F, 0x0A, 0x00,
	}
	_, err = si.Execute(q)
	if err != nil {
		t.Fatal(err)
	}
}

// check that invalid scripts produce an error
func TestInvalidScripts(t *testing.T) {
	q, si := initEnv()

	si.Input = []byte{
		0x01, 0xAA, 0x06,
	}
	_, err := si.Execute(q)
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
	si.Input = []byte{
		0x25, 0xFF, 0x7F,
	}
	_, err = si.Execute(q)
	if err == nil {
		t.Fatal("expected out of bounds error")
	}
}

// test the "verify" opcode
func TestVerify(t *testing.T) {
	q, si := initEnv()

	// generate public key
	publicKey, secretKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	gobPKey, err := publicKey.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	pkeyLen := make([]byte, 2)
	pkeyLen[0] = byte(len(gobPKey))
	pkeyLen[1] = byte(len(gobPKey) >> 8)
	encPKey := append(pkeyLen, gobPKey...)

	// generate signed message
	message := []byte("test")
	signedMessage, err := secretKey.Sign(message)
	if err != nil {
		t.Fatal(err)
	}
	encSm, err := signedMessage.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// construct script
	si.Input = []byte{
		0x25, 0x0F, 0x00, // move data pointer to start of public key
		0x2D, 0x01, //       copy public key into buffer 1
		0x2E, 0x02, //       copy signed message into buffer 2
		0x34, 0x01, 0x02, // verify signature
		0x36, 0x02, 0x00, // if verified, jump over rejection
		0x30, //             reject input
		0xFF, //             terminate script
	}
	si.Input = append(si.Input, encPKey...)
	si.Input = append(si.Input, encSm...)

	// execute script
	_, err = si.Execute(q)
	if err != nil {
		t.Fatal(err)
	}
}
