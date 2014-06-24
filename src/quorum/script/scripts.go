package script

import (
	"siacrypto"
	"siaencoding"
)

func short(length int) (l, h byte) {
	l = byte(length)
	h = byte(length >> 8)
	return
}

func appendAll(slices ...[]byte) []byte {
	var length int
	for _, s := range slices {
		length += len(s)
	}
	all := make([]byte, length)
	var i int
	for _, s := range slices {
		i += copy(all[i:], s)
	}
	return all
}

// the default script
// verifies public key, then transfers control to the input
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	klen, _ := short(siacrypto.PublicKeySize)
	return append([]byte{
		0x26, 0x0D, 0x00, // 00 move data pointer to public key
		0x39, klen, 0x01, // 03 read public key into buffer 1
		0x2D, 0x02, //       06 read signed message into buffer 2
		0x34, 0x01, 0x02, // 08 verify signature
		0x38, //             11 if invalid signature, reject
		0x2F, //             12 execute input
	}, publicKey[:]...)
}

// the bootstrapping script
// creates a wallet with a provided ID and script
var BootstrapScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (wallet id)
	0x01, 0x00, //       02 push 1 (high balance)
	0x02, 0xA8, 0x61, // 04 push 100 (low balance)
	0x2E, 0x01, //       08 read rest of input into buffer 1
	0x32, 0x01, //       10 call create wallet
	0xFF, //             12 exit
}

var TransactionScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (id)
	0x27, 0x08, //       02 push 8 bytes of input (high balance)
	0x27, 0x08, //       04 push 8 bytes of input (low balance)
	0x33, //             06 call send
	0xFF, //             07 exit
}

func CreateWalletInput(walletID uint64, s []byte) []byte {
	id := siaencoding.EncUint64(walletID)
	return append(id, s...)
}

func AddSiblingInput(encSm, encSibling []byte) []byte {
	lenl, lenh := short(len(encSm))
	return appendAll(
		[]byte{lenl, lenh},
		encSm,
		[]byte{
			0x25, 0x08, 0x00, // move data pointer to encoded sibling
			0x2E, 0x01, //       read sibling into buffer 1
			0x31, 0x01, //       call AddSibling
			0xFF, //             exit
		},
		encSibling,
	)
}

func TransactionInput(dst, high, low uint64) []byte {
	return appendAll(
		siaencoding.EncUint64(dst),
		siaencoding.EncUint64(high),
		siaencoding.EncUint64(low),
	)
}

func ResizeSectorEraseInput(atoms, m byte) []byte {
	return []byte{
		0x3A, atoms, m, // simple as that
	}
}

func ProposeUploadInput(encUA []byte) []byte {
	return append([]byte{
		0x25, 0x08, 0x00, // move data pointer to encoded args
		0x2E, 0x01, //       read args into buffer 1
		0x3B, 0x01, //       call ProposeUpload
		0xFF, //             exit
	}, encUA...)
}
