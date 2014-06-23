package script

import (
	"fmt"
)

const hextable = "0123456789ABCDEF"

func encodeHex(b []byte) string {
	dst := make([]byte, 3*len(b))
	for i, v := range b {
		dst[i*3] = hextable[v>>4]
		dst[i*3+1] = hextable[v&0x0F]
		dst[i*3+2] = ' '
	}
	return string(dst[:len(dst)-1])
}

// not perfect, but usually good enough
func findDataSection(script []byte) (index int) {
	index = len(script)
	for i, b := range script {
		if b == 0xFF || b == 0x2F || b == 0x30 || b == 0x38 {
			index = i + 1
		}
		// these are good indicators that we're inside a data block
		if int(b) > len(opTable) || i+opTable[b].argBytes >= len(script) {
			break
		}
	}
	return
}

// this might be added as a field in the instruction type later
var shortArg [256]bool

func init() {
	shortArg[0x02] = true
	shortArg[0x1F] = true
	shortArg[0x20] = true
	shortArg[0x25] = true
	shortArg[0x26] = true
	shortArg[0x36] = true
	shortArg[0x37] = true
}

func BytesToWords(script []byte) (s string) {
	fmt.Println(script)
	dataIndex := findDataSection(script)
	for i := range script {
		if i == dataIndex {
			s += "<-- data section -->\n"
			for i < len(script) {
				b := make([]byte, 32)
				n := copy(b, script[i:])
				s += encodeHex(b[:n]) + "\n"
				i += n
			}
			break
		}
		if script[i] == 0xFF {
			s += "terminate\n"
			continue
		}
		fmt.Println(script[i])
		op := opTable[script[i]]
		s += op.name
		if op.argBytes == 1 {
			s += " " + fmt.Sprint(script[i+1])
		} else if op.argBytes == 2 {
			if shortArg[script[i]] {
				s += " " + fmt.Sprint(s2i(script[i+1], script[i+2]))
			} else {
				s += " " + fmt.Sprint(script[i+1], script[i+2])
			}
		}
		s += "\n"
		i += op.argBytes
	}
	return
}
