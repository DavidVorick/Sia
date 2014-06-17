package script

// the bootstrapping script
// accepts two types of input:
// - run script:    0x00 followed by key
// - create wallet: 0x01 followed by wallet ID and script
// - add sibling:   0x02 followed by encoded sibling
var BootstrapScript = []byte{
	0x27, 0x01, //       00 load first byte of input
	0x04, 0x04, //       02 dup input byte twice
	0x01, 0x00, 0x16, // 04 push 0 and compare
	0x1F, 0x00, 0x0E, // 07 if true, goto 23
	0x01, 0x01, 0x16, // 10 push 1 and compare
	0x1F, 0x00, 0x09, // 13 if true, goto 24
	0x01, 0x02, 0x16, // 16 push 2 and compare
	0x1F, 0x00, 0x0E, // 19 if true, goto 35
	0xFF, //             22 else, exit

	0x2F, //             23 move instruction pointer to input

	0x01, 0x00, //       24 push 0
	0x01, 0x64, //       26 push 100
	0x27, 0x08, //       28 push 8 bytes of input
	0x2E, 0x01, //       30 read rest of input into buffer 1
	0x32, 0x01, //       32 call create wallet
	0xFF, //             34 exit

	0x2E, 0x01, //       35 read rest of input into buffer 1
	0x31, 0x01, //       37 call add sibling
}

// these may be changed to functions later

var CreateWalletInput = []byte{
	0x01, //             00 0x01 byte indicates this is a bootstrap request
}

var AddSiblingInput = []byte{
	0x02, //             00 0x00 byte indicates this is a run script request
}

var DefaultScript = []byte{
	0x2F,
}
