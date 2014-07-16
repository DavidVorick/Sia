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

func SignInput(secretKey *siacrypto.SecretKey, input []byte) (gobSm []byte, err error) {
	sm, err := secretKey.Sign(input)
	if err != nil {
		return
	}
	gobSm, err = sm.GobEncode()
	return
}

func CreateWalletInput(walletID uint64, s []byte) []byte {
	return append(siaencoding.EncUint64(walletID), s...)
}

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

func TransactionInput(dst, high, low uint64) []byte {
	return appendAll(
		[]byte{
			0x25, 0x0B, 0x00, // move data pointer to dst
			0x27, 0x08, //       push 8 bytes of input (id)
			0x27, 0x08, //       push 8 bytes of input (high balance)
			0x27, 0x08, //       push 8 bytes of input (low balance)
			0x33, //             call send
			0xFF, //             exit
		},
		siaencoding.EncUint64(dst),
		siaencoding.EncUint64(high),
		siaencoding.EncUint64(low),
	)
}

func ResizeSectorEraseInput(atoms uint16, m byte) []byte {
	l, h := short(int(atoms))
	return []byte{
		0x02, l, h, // push number of atoms
		0x3A, m, //    call resize
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
