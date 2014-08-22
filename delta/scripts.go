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
func SignScriptInput(si *state.ScriptInput, secretKey siacrypto.SecretKey) (err error) {
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
func AddSiblingInput(wid state.WalletID, deadline uint32, sib state.Sibling, sk siacrypto.SecretKey) (si state.ScriptInput, err error) {
	si.WalletID = wid
	encSib, err := siaencoding.Marshal(sib)
	if err != nil {
		return
	}
	si.Input = appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to encoded sibling
			0xE4, //       push encoded sibling
			0x41, //       call AddSibling
			0xFF, //       exit
		},
		encSib,
	)
	si.Deadline = deadline

	err = SignScriptInput(&si, sk)
	if err != nil {
		return
	}
	return
}

// TransactionInput returns a script that calls the Send function. It is
// intended to be passed to a script that transfers execution to the input.
func SendCoinInput(dest state.WalletID, amount state.Balance) []byte {
	return appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to dest
			0x34, 0x08, // push dest
			0x34, 0x10, // push balance
			0x43, //       call Send
			0xFF, //       exit
		},
		siaencoding.EncUint64(uint64(dest)),
		amount[:],
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
		0xE6, 0xFF, // move data pointer to encoded args
		0xE4, //       push args
		0x45, //       call ProposeUpload
		0xFF, //       exit
	}, encUA...)
}

// The FountainScript acts as a fountain, and can be called to spawn a new
// wallet with a minimal starting balance that will the economy to get flowing.
var FountainScript = []byte{
	0x34, 0x08, //       00 push 8 bytes of input (wallet id)
	0x02, 0xA8, 0x61, // 02 push 25000 (balance)
	0xE4, //             05 push script
	0x42, //             06 call CreateWallet
	0xFF, //             07 exit
}

// DefaultScript returns a script that verifies a signature, and transfers
// control to the input if the verification was successful.
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	keyl, sigl := byte(siacrypto.PublicKeySize), byte(siacrypto.SignatureSize)
	return append([]byte{
		0xE6, 0x38, //       00 move data pointer to public key
		0x34, keyl, //       02 push public key
		0xE4,             // 04 push signed input
		0x40,             // 05 verify signature
		0xE5,             // 06 if invalid signature, reject
		0x33, sigl, 0x00, // 07 move data pointer to input body
		0x38, //             10 execute input
	}, publicKey[:]...)
}
