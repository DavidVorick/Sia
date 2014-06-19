package script

import (
	"siaencoding"
)

// the default script
// for now, this just moves control to the input
// eventually it should allow itself to be overwritten
var DefaultScript = []byte{
	0x2F,
}

// the bootstrapping script
// accepts two types of input:
// - run script:    0x00 followed by key
// - create wallet: 0x01 followed by wallet ID and script
// - add sibling:   0x02 followed by encoded sibling
var BootstrapScript = []byte{
	0x27, 0x01, //       00 load first byte of input
	0x35, 0x00, 0x0C, // 02 if byte == 0, goto 12
	0x35, 0x01, 0x0D, // 05 if byte == 1, goto 13
	0x35, 0x02, 0x18, // 08 if byte == 2, goto 24
	0xFF, //             11 else, exit

	0x2F, //             12 move instruction pointer to input

	0x27, 0x08, //       13 push 8 bytes of input (wallet id)
	0x01, 0x00, //       15 push 0 (high balance)
	0x01, 0x64, //       17 push 100 (low balance)
	0x2E, 0x01, //       19 read rest of input into buffer 1
	0x32, 0x01, //       21 call create wallet
	0xFF, //             23 exit

	0x2E, 0x01, //       24 read rest of input into buffer 1
	0x31, 0x01, //       26 call add sibling
	0xFF, //             28 exit
}

func CreateWalletInput(walletID uint64, s []byte) []byte {
	id := siaencoding.EncUint64(walletID)
	return append([]byte{0x01}, append(id, s...)...)
}

func AddSiblingInput(encSib []byte) []byte {
	return append([]byte{0x02}, encSib...)
}

var TransactionScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (id)
	0x27, 0x08, //       02 push 8 bytes of input (high balance)
	0x27, 0x08, //       04 push 8 bytes of input (low balance)
	0x33, //             06 call send
	0xFF, //             07 exit
}

func TransactionInput(dst, high, low uint64) []byte {
	wallet := siaencoding.EncUint64(dst)
	amount := append(
		siaencoding.EncUint64(high),
		siaencoding.EncUint64(low)...,
	)
	return append(wallet, amount...)
}
