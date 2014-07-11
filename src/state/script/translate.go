package script

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
		if int(b) >= len(opTable) || i+opTable[b].argBytes >= len(script) {
			break
		}
	}
	return
}

// this might be added as a field in the instruction type later
var shortArg [256]bool

var opcodeMap map[string]byte

func init() {
	shortArg[0x02] = true
	shortArg[0x1F] = true
	shortArg[0x20] = true
	shortArg[0x25] = true
	shortArg[0x26] = true
	shortArg[0x36] = true
	shortArg[0x37] = true

	// build name -> opcode map
	opcodeMap = make(map[string]byte)
	for opcode, op := range opTable {
		opcodeMap[op.name] = opcode
	}
}

func BytesToWords(script []byte) (s string, err error) {
	// locate data section, if there is one
	dataIndex := findDataSection(script)

	for i := 0; i < len(script); i++ {
		if i == dataIndex {
			s += "<--data-->\n"
			// print hex-formatted data in rows of 32
			for i < len(script) {
				b := make([]byte, 32)
				n := copy(b, script[i:])
				s += encodeHex(b[:n]) + "\n"
				i += n
			}
			break
		}

		// unknown opcode
		if int(script[i]) >= len(opTable) {
			err = errors.New("error parsing script")
			return
		}

		op := opTable[script[i]]
		s += op.name

		// unrolled loop, since there are only two arguments max
		if op.argBytes == 1 {
			s += fmt.Sprint(" ", script[i+1])
		} else if op.argBytes == 2 {
			// combine two bytes into one number where appropriate
			if shortArg[script[i]] {
				s += fmt.Sprint(" ", s2i(script[i+1], script[i+2]))
			} else {
				s += fmt.Sprint(" ", script[i+1], script[i+2])
			}
		}

		s += "\n"
		i += op.argBytes
	}

	return
}

func WordsToBytes(script string) (b []byte, err error) {
	// simple tokenization using a whitespace separator
	tokens := strings.Fields(script)

	for i := 0; i < len(tokens); i++ {
		// parse rest of script as hex literals
		if tokens[i] == "<--data-->" {
			var data byte
			for i++; i < len(tokens); i++ {
				fmt.Sscanf(tokens[i], "%X", &data)
				b = append(b, data)
			}
			return
		}

		// parse opcode
		opcode, ok := opcodeMap[tokens[i]]
		if !ok {
			err = errors.New(fmt.Sprint("expected opcode, got ", tokens[i]))
			return
		}
		b = append(b, opcode)

		// parse argument(s)
		numArgs := opTable[opcode].argBytes
		if shortArg[opcode] {
			numArgs = 1
		}
		if i+numArgs > len(tokens) {
			err = errors.New(fmt.Sprint("not enough arguments to ", tokens[i]))
			return
		}
		for j := 1; j <= numArgs; j++ {
			arg, convErr := strconv.Atoi(tokens[i+j])
			if convErr != nil {
				err = errors.New(fmt.Sprintf("invalid argument \"%s\" to opcode %s", tokens[i+j], tokens[i]))
				return
			}
			// convert single number to two bytes
			if shortArg[opcode] {
				if arg > 0xFFFF {
					err = errors.New("argument overflows short")
					return
				}
				b = append(b, byte(arg), byte(arg>>8))
			} else {
				if arg > 0xFF {
					err = errors.New("argument overflows byte")
					return
				}
				b = append(b, byte(arg))
			}
		}

		i += numArgs
	}

	return
}
