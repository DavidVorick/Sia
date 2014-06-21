package script

import (
	"siaencoding"
)

func shortLen(data []byte) (l, h byte) {
	l = byte(len(data))
	h = byte(len(data) >> 8)
	return
}

// the default script
// verifies public key, then transfers control to the input
func DefaultScript(encPKey []byte) []byte {
	lenl, lenh := shortLen(encPKey)
	negl, negh := 0xFF-lenl, 0xFF-lenh
	return append([]byte{
		0x25, negl, negh, // 00 move data pointer to public key
		0x02, lenl, lenh, // 03 push length of public key
		0x2B, 0x01, //       06 read public key into buffer 1
		0x2D, 0x02, //       08 read signed message into buffer 2
		0x34, 0x01, 0x02, // 10 verify signature
		0x1F, 0x01, 0x01, // 13 if verified, goto 17
		0x30, //             16 else, reject input
		0x2F, //             17 execute input
	}, encPKey...)
}

// the bootstrapping script
// creates a wallet with a provided ID and script
var BootstrapScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (wallet id)
	0x01, 0x00, //       02 push 0 (high balance)
	0x01, 0x64, //       04 push 100 (low balance)
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
	lenh, lenl := shortLen(encSm)
	s := append([]byte{lenh, lenl}, encSm...)
	s = append(s, []byte{
		0x25, 0x08, 0x00, // move data pointer to encoded sibling
		0x2E, 0x01, //       read sibling into buffer 1
		0x31, 0x01, //       call add sibling
		0xFF, //             exit
	}...)
	return append(s, encSibling...)
}

func TransactionInput(dst, high, low uint64) []byte {
	wallet := siaencoding.EncUint64(dst)
	amount := append(
		siaencoding.EncUint64(high),
		siaencoding.EncUint64(low)...,
	)
	return append(wallet, amount...)
}
