package script

import (
	"bytes"
	"encoding/gob"
	"errors"
	"network"
	"quorum"
	"reflect"
	"siacrypto"
	"siaencoding"
)

var opTable = []instruction{
	instruction{0x00, "nop", 0, reflect.ValueOf(op_nop), 1},
	instruction{0x01, "pushb", 1, reflect.ValueOf(op_pushb), 2},
	instruction{0x02, "pushs", 2, reflect.ValueOf(op_pushs), 2},
	instruction{0x03, "pop", 0, reflect.ValueOf(op_pop), 1},
	instruction{0x04, "dup", 0, reflect.ValueOf(op_dup), 2},
	instruction{0x05, "swap", 0, reflect.ValueOf(op_swap), 2},
	instruction{0x06, "addi", 0, reflect.ValueOf(op_addi), 2},
	instruction{0x07, "addf", 0, reflect.ValueOf(op_addf), 3},
	instruction{0x08, "subi", 0, reflect.ValueOf(op_subi), 2},
	instruction{0x09, "subf", 0, reflect.ValueOf(op_subf), 3},
	instruction{0x0A, "muli", 0, reflect.ValueOf(op_muli), 2},
	instruction{0x0B, "mulf", 0, reflect.ValueOf(op_mulf), 3},
	instruction{0x0C, "divi", 0, reflect.ValueOf(op_divi), 2},
	instruction{0x0D, "divf", 0, reflect.ValueOf(op_divf), 3},
	instruction{0x0E, "modi", 0, reflect.ValueOf(op_modi), 3},
	instruction{0x0F, "negi", 0, reflect.ValueOf(op_negi), 2},
	instruction{0x10, "negf", 0, reflect.ValueOf(op_negf), 3},
	instruction{0x11, "bor", 0, reflect.ValueOf(op_bor), 2},
	instruction{0x12, "band", 0, reflect.ValueOf(op_band), 2},
	instruction{0x13, "bxor", 0, reflect.ValueOf(op_bxor), 2},
	instruction{0x14, "shln", 1, reflect.ValueOf(op_shln), 2},
	instruction{0x15, "shrn", 1, reflect.ValueOf(op_shrn), 2},
	instruction{0x16, "eq", 0, reflect.ValueOf(op_eq), 2},
	instruction{0x17, "ne", 0, reflect.ValueOf(op_ne), 2},
	instruction{0x18, "lti", 0, reflect.ValueOf(op_lti), 2},
	instruction{0x19, "ltf", 0, reflect.ValueOf(op_ltf), 2},
	instruction{0x1A, "gti", 0, reflect.ValueOf(op_gti), 2},
	instruction{0x1B, "gtf", 0, reflect.ValueOf(op_gtf), 2},
	instruction{0x1C, "lnot", 0, reflect.ValueOf(op_lnot), 2},
	instruction{0x1D, "lor", 0, reflect.ValueOf(op_lor), 2},
	instruction{0x1E, "land", 0, reflect.ValueOf(op_land), 2},
	instruction{0x1F, "if", 2, reflect.ValueOf(op_if), 2},
	instruction{0x20, "goto", 2, reflect.ValueOf(op_goto), 1},
	instruction{0x21, "regs", 1, reflect.ValueOf(op_regs), 2},
	instruction{0x22, "regl", 1, reflect.ValueOf(op_regl), 2},
	instruction{0x23, "inci", 1, reflect.ValueOf(op_inci), 2},
	instruction{0x24, "deci", 1, reflect.ValueOf(op_deci), 2},
	instruction{0x25, "dmov", 2, reflect.ValueOf(op_dmov), 1},
	instruction{0x26, "dgoto", 2, reflect.ValueOf(op_dgoto), 1},
	instruction{0x27, "dpush", 1, reflect.ValueOf(op_dpush), 2},
	instruction{0x28, "dregs", 2, reflect.ValueOf(op_dregs), 2},
	instruction{0x29, "repb", 0, reflect.ValueOf(op_repb), 2},
	instruction{0x2A, "reps", 0, reflect.ValueOf(op_reps), 2},
	instruction{0x2B, "bufc", 2, reflect.ValueOf(op_bufc), 2},
	instruction{0x2C, "bufp", 2, reflect.ValueOf(op_bufp), 2},
	instruction{0x2D, "xfer", 0, reflect.ValueOf(op_xfer), 1},
	instruction{0x2E, "rej", 0, reflect.ValueOf(op_rej), 0},
	instruction{0x2F, "asib", 2, reflect.ValueOf(op_asib), 5},
}

// helper functions
func v2i(b value) int64 {
	return siaencoding.DecInt64(b[:])
}

func i2v(i int64) (v value) {
	b := siaencoding.EncInt64(i)
	copy(v[:], b)
	return
}

func v2f(b value) float64 {
	return siaencoding.DecFloat64(b[:])
}

func f2v(f float64) (v value) {
	b := siaencoding.EncFloat64(f)
	copy(v[:], b)
	return
}

// convert two bytes to signed short
func s2i(high, low byte) int {
	return int(int16(high)<<8 + int16(low))
}

func b2y(b bool) byte {
	if b {
		return 0x01
	} else {
		return 0x00
	}
}

func y2b(b value) bool {
	return v2i(b) != 0
}

// opcodes

func op_nop() (err error) {
	return
}

func op_pushb(b byte) (err error) {
	err = push(value{b})
	return
}

func op_pushs(h, l byte) (err error) {
	err = push(value{l, h})
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
	err = push(a)
	return
}

func op_swap() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	push(a)
	err = push(b)
	return
}

func op_addi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) + v2i(b)))
	return
}

func op_addf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) + v2f(b)))
	return
}

func op_subi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) - v2i(b)))
	return
}

func op_subf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) - v2f(b)))
	return
}

func op_muli() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) * v2i(b)))
	return
}

func op_mulf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) * v2f(b)))
	return
}

func op_divi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	if v2i(b) == 0 {
		return errors.New("divide by zero")
	}
	err = push(i2v(v2i(a) / v2i(b)))
	return
}

func op_divf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	if v2f(b) == 0.0 {
		return errors.New("divide by zero")
	}
	err = push(f2v(v2f(a) / v2f(b)))
	return
}

func op_modi() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) % v2i(b)))
	return
}

func op_negi() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(-v2i(a)))
	return
}

func op_negf() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(-v2f(a)))
	return
}

func op_bor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) | v2i(b)))
	return
}

func op_band() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) & v2i(b)))
	return
}

func op_bxor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) ^ v2i(b)))
	return
}

func op_shln(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) << n))
	return
}

func op_shrn(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) >> n))
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
	op_pushb(b2y(v2i(a) < v2i(b)))
	return
}

func op_ltf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(v2f(a) < v2f(b)))
	return
}

func op_gti() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(v2i(a) > v2i(b)))
	return
}

func op_gtf() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_pushb(b2y(v2f(a) > v2f(b)))
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
		err = op_goto(offh, offl)
	}
	return
}

func op_goto(offh, offl byte) (err error) {
	iptr += s2i(offh, offl)
	if iptr < 0 {
		err = errors.New("jumped to invalid index")
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
	err = push(registers[reg])
	return
}

func op_inci(reg, n byte) (err error) {
	registers[reg] = i2v(v2i(registers[reg]) + int64(n))
	return
}

func op_deci(reg, n byte) (err error) {
	registers[reg] = i2v(v2i(registers[reg]) - int64(n))
	return
}

func op_dmov(loch, locl byte) (err error) {
	dptr += s2i(loch, locl)
	if dptr < 0 || dptr > len(script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_dgoto(loch, locl byte) (err error) {
	dptr = s2i(loch, locl)
	if dptr < 0 || dptr > len(script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_dpush(n byte) (err error) {
	var v value
	b := make([]byte, n)
	copy(b, script[dptr:])
	copy(v[:], b)
	err = op_push(v)
	return
}

func op_dregs(n, reg byte) (err error) {
	var v value
	b := make([]byte, n)
	copy(b, script[dptr:])
	copy(v[:], b)
	registers[reg] = v
	return
}

func op_repb() (err error) {
	err, a := op_pop()
	script[dptr] = a[0]
	return
}

func op_reps() (err error) {
	err, a := op_pop()
	script[dptr] = a[0]
	script[dptr+1] = a[1]
	return
}

func op_bufc(lenh, lenl byte) (err error) {
	length := s2i(lenh, lenl)
	buffer = make([]byte, length)
	copy(buffer, script[dptr:])
	return
}

func op_bufp(lenh, lenl byte) (err error) {
	length := s2i(lenh, lenl)
	// extend script if necessary
	if dptr+length > len(script) {
		ext := make([]byte, dptr+length-len(script))
		script = append(script, ext...)
	}
	b := make([]byte, length)
	copy(b, buffer)
	copy(script[dptr:], b)
	return
}

func op_xfer() (err error) {
	iptr = dptr
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

func op_awall() (err error) {
	_, id := op_pop()
	_, lbal := op_pop()
	err, ubal := op_pop()
	if err != nil {
		return
	}
	newscript := buffer
	atoms := len(newscript)/4096 + 1
	q.CreateWallet(wallet, id, ubal, lbal, atoms, newscript)
	return
}

func op_send() (err error) {
	_, id := op_pop()
	_, lbal := op_pop()
	err, ubal := op_pop()
	if err != nil {
		return
	}
	q.Send(wallet, ubal, lbal, id)
	return
}
