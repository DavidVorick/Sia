package delta

import (
	"siacrypto"
	"siaencoding"
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

// SignInput takes a secret key and an input, and returns the signature of the
// input followed by input itself.
func SignInput(secretKey siacrypto.SecretKey, input []byte) ([]byte, error) {
	sig, err := secretKey.Sign(input)
	if err != nil {
		return nil, err
	}
	return append(sig[:], input...), nil
}

// CreateWalletInput returns the appropriate input for the CreateWallet script.
func CreateWalletInput(walletID uint64, s []byte) []byte {
	return append(siaencoding.EncUint64(walletID), s...)
}

// AddSiblingInput returns a script that calls the AddSibling function. It is
// intended to be passed to a script that transfers execution to the input.
func AddSiblingInput(encSibling []byte) []byte {
	return appendAll(
		[]byte{
			0x25, 0x08, 0x00, // move data pointer to encoded sibling
			0x2E, 0x01, //       read sibling into buffer 1
			0x31, 0x01, //       call AddSibling
			0xFF, //             exit
		},
		encSibling,
	)
}

// TransactionInput returns a script that calls the Send function. It is
// intended to be passed to a script that transfers execution to the input.
func TransactionInput(dst, high, low uint64) []byte {
	return appendAll(
		[]byte{
			0x33, 0x0B, 0x00, // move data pointer to dst
			0x34, 0x08, //       push 8 bytes of input (id)
			0x34, 0x08, //       push 8 bytes of input (high balance)
			0x34, 0x08, //       push 8 bytes of input (low balance)
			0x43, //             call Send
			0xFF, //             exit
		},
		siaencoding.EncUint64(dst),
		siaencoding.EncUint64(high),
		siaencoding.EncUint64(low),
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
		0xE2, 0x01, //       read args into buffer 1
		0x45, 0x01, //       call ProposeUpload
		0xFF, //             exit
	}, encUA...)
}

// The BootstrapScript acts as a fountain, and can be called to spawn a new
// wallet with a minimal starting balance that will the economy to get flowing.
var BootstrapScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (wallet id)
	0x01, 0x00, //       02 push 1 (high balance)
	0x02, 0xA8, 0x61, // 04 push 100 (low balance)
	0x2E, 0x01, //       08 read rest of input into buffer 1
	0x32, 0x01, //       10 call create wallet
	0xFF, //             12 exit
}

// DefaultScript returns a script that verifies a signature, and transfers
// control to the input if the verification was successful.
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	return append([]byte{
		0x26, 0x10, 0x00, // 00 move data pointer to public key
		0x39, 0x20, 0x01, // 03 read public key into buffer 1
		0x2E, 0x02, //       06 read signed message into buffer 2
		0x34, 0x01, 0x02, // 08 verify signature
		0x38,             // 11 if invalid signature, reject
		0x26, 0x70, 0x00, // 12 move data pointer to input body
		0x2F, //             15 execute input
	}, publicKey[:]...)
}
