package script

import (
	"bytes"
	"encoding/gob"
	"errors"
	"network"
	"quorum"
	"reflect"
	"siacrypto"
	"unsafe"
)

var opTable = []instruction{
	instruction{0x00, 0, reflect.ValueOf(op_nop), 1},
	instruction{0x01, 1, reflect.ValueOf(op_pushb), 2},
	instruction{0x02, 2, reflect.ValueOf(op_pushs), 2},
	instruction{0x03, 0, reflect.ValueOf(op_pop), 1},
	instruction{0x04, 0, reflect.ValueOf(op_dup), 2},
	instruction{0x05, 0, reflect.ValueOf(op_swap), 2},
	instruction{0x06, 0, reflect.ValueOf(op_addi), 2},
	instruction{0x07, 0, reflect.ValueOf(op_subi), 2},
	instruction{0x08, 0, reflect.ValueOf(op_muli), 2},
	instruction{0x09, 0, reflect.ValueOf(op_divi), 2},
	instruction{0x0A, 0, reflect.ValueOf(op_modi), 3},
	instruction{0x0B, 0, reflect.ValueOf(op_negi), 2},
	instruction{0x0C, 0, reflect.ValueOf(op_bor), 2},
	instruction{0x0D, 0, reflect.ValueOf(op_band), 2},
	instruction{0x0E, 0, reflect.ValueOf(op_bxor), 2},
	instruction{0x0F, 1, reflect.ValueOf(op_shln), 2},
	instruction{0x10, 1, reflect.ValueOf(op_shrn), 2},
	instruction{0x11, 0, reflect.ValueOf(op_eq), 2},
	instruction{0x12, 0, reflect.ValueOf(op_ne), 2},
	instruction{0x13, 0, reflect.ValueOf(op_lti), 2},
	instruction{0x14, 0, reflect.ValueOf(op_gti), 2},
	instruction{0x15, 0, reflect.ValueOf(op_lnot), 2},
	instruction{0x16, 0, reflect.ValueOf(op_lor), 2},
	instruction{0x17, 0, reflect.ValueOf(op_land), 2},
	instruction{0x18, 2, reflect.ValueOf(op_if), 2},
	instruction{0x19, 2, reflect.ValueOf(op_goto), 1},
	instruction{0x1A, 1, reflect.ValueOf(op_regs), 2},
	instruction{0x1B, 1, reflect.ValueOf(op_regl), 2},
	instruction{0x1C, 1, reflect.ValueOf(op_inci), 2},
	instruction{0x1D, 1, reflect.ValueOf(op_deci), 2},
	instruction{0x1E, 2, reflect.ValueOf(op_blks), 2},
	instruction{0x1F, 2, reflect.ValueOf(op_blkl), 2},
	instruction{0x20, 0, reflect.ValueOf(op_rej), 0},
	instruction{0x21, 2, reflect.ValueOf(op_asib), 5},
}

// helper functions
func y2i(b value) int64 {
	return *(*int64)(unsafe.Pointer(&b))
}

func i2y(i int64) value {
	return *(*value)(unsafe.Pointer(&i))
}

func s2i(high, low byte) int {
	return int((high << 8) + low)
}

func b2y(b bool) byte {
	if b {
		return 0x01
	} else {
		return 0x00
	}
}

func y2b(b value) bool {
	return y2i(b) != 0
}

// opcodes

func op_nop() (err error) {
	return
}

func op_pushb(b byte) (err error) {
	push(value{b})
	return
}

func op_pushs(h, l byte) (err error) {
	push(value{l, h})
	return
}

func op_pop() (err error, val value) {
	if stackLen < 1 {
		err = errors.New("stack empty")
		return
	}
	val = stack.val
	stack = stack.next
	stackLen--
	return
}

func op_dup() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	push(a)
	push(a)
	return
}

func op_swap() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(a)
	push(b)
	return
}

func op_addi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) + y2i(b)))
	return
}

func op_subi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) - y2i(b)))
	return
}

func op_muli() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) * y2i(b)))
	return
}

func op_divi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) / y2i(b)))
	return
}

func op_modi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) % y2i(b)))
	return
}

func op_negi() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	push(i2y(-y2i(a)))
	return
}

func op_bor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) | y2i(b)))
	return
}

func op_band() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) & y2i(b)))
	return
}

func op_bxor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) ^ y2i(b)))
	return
}

func op_shln(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) << n))
	return
}

func op_shrn(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	push(i2y(y2i(a) >> n))
	return
}

func op_eq() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(a == b))
	return
}

func op_ne() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(a != b))
	return
}

func op_lti() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(y2i(a) < y2i(b)))
	return
}

func op_gti() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(y2i(a) > y2i(b)))
	return
}

func op_lnot() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(!y2b(a)))
	return
}

func op_lor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(y2b(a) || y2b(b)))
	return
}

func op_land() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(y2b(a) && y2b(b)))
	return
}

func op_if(offh, offl byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	if y2b(a) {
		return op_goto(offh, offl)
	}
	return
}

func op_goto(offh, offl byte) (err error) {
	iptr += s2i(offh, offl)
	if iptr < 0 {
		return errors.New("jumped to invalid index")
	}
	// the iptr > len(script) case is handled inside Execute
	return
}

func op_regs(reg byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	registers[reg] = a
	return
}

func op_regl(reg byte) (err error) {
	push(registers[reg])
	return
}

func op_inci(reg, n byte) (err error) {
	registers[reg] = i2y(y2i(registers[reg]) + int64(n))
	return
}

func op_deci(reg, n byte) (err error) {
	registers[reg] = i2y(y2i(registers[reg]) - int64(n))
	return
}

func op_blks(loch, locl byte) (err error) {
	err, a := op_pop()
	addr := s2i(loch, locl)
	if addr < 0 || addr+8 > len(script) {
		return errors.New("invalid data access")
	}
	copy(script[addr:addr+8], a[:])
	return
}

func op_blkl(loch, locl byte) (err error) {
	addr := s2i(loch, locl)
	if addr < 0 || addr+8 > len(script) {
		return errors.New("invalid data access")
	}
	var a value
	copy(a[:], script[addr:addr+8])
	push(a)
	return
}

func op_rej() (err error) {
	return errors.New("rejected input")
}

func op_asib(loc byte, length byte) (err error) {
	// read encoded sibling from data block
	if int(loc+length) > len(script) {
		return errors.New("invalid data access")
	}
	encSibling := script[loc : loc+length]

	// decode sibling
	var address network.Address
	var key siacrypto.PublicKey
	reader := bytes.NewBuffer(encSibling)
	decoder := gob.NewDecoder(reader)
	decoder.Decode(&address)
	decoder.Decode(&key)

	// add sibling
	q.AddSibling(quorum.NewSibling(address, &key))
	return
}
