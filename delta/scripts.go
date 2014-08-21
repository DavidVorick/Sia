package delta

import (
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
)

func short(length int) (l, h byte) {
	l = byte(length)
	h = byte(length >> 8)
	return
}

// why isn't this a builtin? Definitely nicer than bytes.Join
func appendAll(slices ...[]byte) []byte {
	var length int
	for _, s := range slices {
		length += len(s)
	}
	all := make([]byte, length)
	i := 0
	for _, s := range slices {
		i += copy(all[i:], s)
	}
	return all
}

// Sign takes a SecretKey and modifies the receiving ScriptInput to contain a
// signature of its own data. Currently, only the Input field is included in
// the signature.
func (si *ScriptInput) Sign(secretKey siacrypto.SecretKey) (err error) {
	sig, err := secretKey.Sign(si.Input)
	if err != nil {
		return
	}
	si.Input = append(sig[:], si.Input...)
	return
}

// CreateWalletInput returns a script that can be used by the fountain wallet
// to create a new wallet.
func CreateFountainWalletInput(id state.WalletID, s []byte) []byte {
	return append(siaencoding.EncUint64(uint64(id)), s...)
}

// AddSiblingInput returns a signed ScriptInput that calls the AddSibling
// function. It is intended to be passed to a script that transfers execution
// to the input.
func AddSiblingInput(wid state.WalletID /*, deadline uint64*/, sib state.Sibling, sk siacrypto.SecretKey) (si ScriptInput, err error) {
	si.WalletID = wid
	encSib, err := siaencoding.Marshal(sib)
	if err != nil {
		return
	}
	si.Input = appendAll(
		[]byte{
			0x33, 0x06, 0x00, // move data pointer to encoded sibling
			0xE4, //             push encoded sibling
			0x41, //             call AddSibling
			0xFF, //             exit
		},
		encSib,
	)
	//si.Deadline = deadline

	err = si.Sign(sk)
	if err != nil {
		return
	}
	return
}

// TransactionInput returns a script that calls the Send function. It is
// intended to be passed to a script that transfers execution to the input.
func SendCoinInput(destination state.WalletID, amount state.Balance) []byte {
	return appendAll(
		[]byte{
			0x33, 0x0B, 0x00, // move data pointer to dst
			0x34, 0x08, //       push 8 bytes of input (id)
			0x34, 0x08, //       push 8 bytes of input (high balance)
			0x34, 0x08, //       push 8 bytes of input (low balance)
			0x43, //             call Send
			0xFF, //             exit
		},
		siaencoding.EncUint64(uint64(destination)),
		//siaencoding.EncUint64(high),
		//siaencoding.EncUint64(low),
	)
}

// ResizeSectorEraseInput returns a script that calls the ResizeSectorErase
// function. It is intended to be passed to a script that transfers execution
// to the input.
func ResizeSectorEraseInput(atoms uint16, m byte) []byte {
	l, h := short(int(atoms))
	return []byte{
		0x02, l, h, // push number of atoms
		0x44, m, //    call resize
	}
}

// ProposeUploadInput returns a script that calls the ProposeUploadInput
// function. It is intended to be passed to a script that transfers execution
// to the input.
func ProposeUploadInput(encUA []byte) []byte {
	return append([]byte{
		0x33, 0x08, 0x00, // move data pointer to encoded args
		0xE4, //             push args
		0x45, //             call ProposeUpload
		0xFF, //             exit
	}, encUA...)
}

// The FountainScript acts as a fountain, and can be called to spawn a new
// wallet with a minimal starting balance that will the economy to get flowing.
var FountainScript = []byte{
	0x34, 0x08, //       00 push 8 bytes of input (wallet id)
	0x02, 0xA8, 0x61, // 02 push 25000 (balance)
	0xE4, //             05 push script
	0x42, //             06 call CreateWallet
	0xFF, //             07 exit (not technically necessary here)
}

// DefaultScript returns a script that verifies a signature, and transfers
// control to the input if the verification was successful.
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	keyl, sigl := byte(siacrypto.PublicKeySize), byte(siacrypto.SignatureSize)
	return append([]byte{
		0x32, 0x0C, 0x00, // 00 move data pointer to public key
		0x34, keyl, //       03 push public key
		0xE4,             // 05 push signed input
		0x40,             // 06 verify signature
		0xE5,             // 07 if invalid signature, reject
		0x33, sigl, 0x00, // 08 move data pointer to input body
		0x38, //             11 execute input
	}, publicKey[:]...)
}
