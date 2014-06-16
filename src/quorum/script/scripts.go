package script

// the bootstrapping script
// accepts two types of input:
// - run script: 0 followed by key
// - create wallet: non-zero followed by script body
var BootstrapScript = []byte{
	0x27, 0x01, //       00 load first byte of input
	0x1F, 0x00, 0x02, // 02 if byte == 0
	0x2F, //             05     move instruction pointer to input
	//                      else
	0x01, 0x01, //       06    push 0
	0x01, 0x64, //       08    push 100
	0x27, 0x08, //       10    push 8 bytes of input
	0x2D, 0x01, //       12    read rest of input into buffer 1
	0x30, 0x01, //       14    call create wallet
}

// Input for the bootstrap script to add a sibling.
// The encoded sibling must be appended to this script before execution,
// e.g. append(script.AddSiblingInput, encSibling...)
var AddSiblingInput = []byte{
	0x00, //               00 zero byte indicates this is a run script request
	//                        (BootstrapScript now moves instruction pointer to 0x25)
	0x25, 0x00, 0x08, //   01 move data pointer to start of encSibling
	0x2E, 0x01, //         04 copy encoded sibling into buffer 1
	0x31, 0x01, //         06 call addSibling on buffer 1
	0xFF, //               08 terminate here to avoid processing the sibling data as opcodes
}
