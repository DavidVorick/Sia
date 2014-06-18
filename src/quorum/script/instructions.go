package script

import (
	"errors"
	"quorum"
	"reflect"
	"siacrypto"
	"siaencoding"
)

var opTable = []instruction{
	instruction{0x00, "no_op", 0, reflect.ValueOf(op_no_op), 1},
	instruction{0x01, "push_byte", 1, reflect.ValueOf(op_push_byte), 2},
	instruction{0x02, "push_short", 2, reflect.ValueOf(op_push_short), 2},
	instruction{0x03, "pop", 0, reflect.ValueOf(op_pop), 1},
	instruction{0x04, "dup", 0, reflect.ValueOf(op_dup), 2},
	instruction{0x05, "swap", 0, reflect.ValueOf(op_swap), 2},
	instruction{0x06, "add_int", 0, reflect.ValueOf(op_add_int), 2},
	instruction{0x07, "add_float", 0, reflect.ValueOf(op_add_float), 3},
	instruction{0x08, "sub_int", 0, reflect.ValueOf(op_sub_int), 2},
	instruction{0x09, "sub_float", 0, reflect.ValueOf(op_sub_float), 3},
	instruction{0x0A, "mul_int", 0, reflect.ValueOf(op_mul_int), 2},
	instruction{0x0B, "mul_float", 0, reflect.ValueOf(op_mul_float), 3},
	instruction{0x0C, "div_int", 0, reflect.ValueOf(op_div_int), 2},
	instruction{0x0D, "div_float", 0, reflect.ValueOf(op_div_float), 3},
	instruction{0x0E, "mod_int", 0, reflect.ValueOf(op_mod_int), 3},
	instruction{0x0F, "neg_int", 0, reflect.ValueOf(op_neg_int), 2},
	instruction{0x10, "neg_float", 0, reflect.ValueOf(op_neg_float), 3},
	instruction{0x11, "binary_or", 0, reflect.ValueOf(op_binary_or), 2},
	instruction{0x12, "binary_and", 0, reflect.ValueOf(op_binary_and), 2},
	instruction{0x13, "binary_xor", 0, reflect.ValueOf(op_binary_xor), 2},
	instruction{0x14, "shift_left", 1, reflect.ValueOf(op_shift_left), 2},
	instruction{0x15, "shift_right", 1, reflect.ValueOf(op_shift_right), 2},
	instruction{0x16, "equal", 0, reflect.ValueOf(op_equal), 2},
	instruction{0x17, "not_equal", 0, reflect.ValueOf(op_not_equal), 2},
	instruction{0x18, "less_int", 0, reflect.ValueOf(op_less_int), 2},
	instruction{0x19, "less_float", 0, reflect.ValueOf(op_less_float), 2},
	instruction{0x1A, "greater_int", 0, reflect.ValueOf(op_greater_int), 2},
	instruction{0x1B, "greater_float", 0, reflect.ValueOf(op_greater_float), 2},
	instruction{0x1C, "logical_not", 0, reflect.ValueOf(op_logical_not), 2},
	instruction{0x1D, "logical_or", 0, reflect.ValueOf(op_logical_or), 2},
	instruction{0x1E, "logical_and", 0, reflect.ValueOf(op_logical_and), 2},
	instruction{0x1F, "if", 2, reflect.ValueOf(op_if), 2},
	instruction{0x20, "goto", 2, reflect.ValueOf(op_goto), 1},
	instruction{0x21, "reg_store", 1, reflect.ValueOf(op_reg_store), 2},
	instruction{0x22, "reg_load", 1, reflect.ValueOf(op_reg_load), 2},
	instruction{0x23, "reg_inc", 1, reflect.ValueOf(op_reg_inc), 2},
	instruction{0x24, "reg_dec", 1, reflect.ValueOf(op_reg_dec), 2},
	instruction{0x25, "data_move", 2, reflect.ValueOf(op_data_move), 1},
	instruction{0x26, "data_goto", 2, reflect.ValueOf(op_data_goto), 1},
	instruction{0x27, "data_push", 1, reflect.ValueOf(op_data_push), 2},
	instruction{0x28, "data_reg", 2, reflect.ValueOf(op_data_reg), 2},
	instruction{0x29, "replace_byte", 0, reflect.ValueOf(op_replace_byte), 2},
	instruction{0x2A, "replace_short", 0, reflect.ValueOf(op_replace_short), 2},
	instruction{0x2B, "buf_copy", 1, reflect.ValueOf(op_buf_copy), 2},
	instruction{0x2C, "buf_paste", 1, reflect.ValueOf(op_buf_paste), 2},
	instruction{0x2D, "buf_prefix", 1, reflect.ValueOf(op_buf_prefix), 2},
	instruction{0x2E, "buf_rest", 1, reflect.ValueOf(op_buf_rest), 2},
	instruction{0x2F, "transfer", 0, reflect.ValueOf(op_transfer), 1},
	instruction{0x30, "reject", 0, reflect.ValueOf(op_reject), 0},
	instruction{0x31, "add_sibling", 1, reflect.ValueOf(op_add_sibling), 5},
	instruction{0x32, "add_wallet", 1, reflect.ValueOf(op_add_wallet), 5},
	instruction{0x33, "send", 0, reflect.ValueOf(op_send), 5},
	instruction{0x34, "verify", 0, reflect.ValueOf(op_verify), 9},
	instruction{0x35, "switch", 2, reflect.ValueOf(op_switch), 3},
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

func op_no_op() (err error) {
	return
}

func op_push_byte(b byte) (err error) {
	err = push(value{b})
	return
}

func op_push_short(h, l byte) (err error) {
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

func op_add_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) + v2i(b)))
	return
}

func op_add_float() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) + v2f(b)))
	return
}

func op_sub_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) - v2i(b)))
	return
}

func op_sub_float() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) - v2f(b)))
	return
}

func op_mul_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) * v2i(b)))
	return
}

func op_mul_float() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) * v2f(b)))
	return
}

func op_div_int() (err error) {
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

func op_div_float() (err error) {
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

func op_mod_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) % v2i(b)))
	return
}

func op_neg_int() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(-v2i(a)))
	return
}

func op_neg_float() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(f2v(-v2f(a)))
	return
}

func op_binary_or() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) | v2i(b)))
	return
}

func op_binary_and() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) & v2i(b)))
	return
}

func op_binary_xor() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) ^ v2i(b)))
	return
}

func op_shift_left(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) << n))
	return
}

func op_shift_right(n byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) >> n))
	return
}

func op_equal() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(a == b))
	return
}

func op_not_equal() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(a != b))
	return
}

func op_less_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(v2i(a) < v2i(b)))
	return
}

func op_less_float() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(v2f(a) < v2f(b)))
	return
}

func op_greater_int() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(v2i(a) > v2i(b)))
	return
}

func op_greater_float() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(v2f(a) > v2f(b)))
	return
}

func op_logical_not() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(!y2b(a)))
	return
}

func op_logical_or() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(y2b(a) || y2b(b)))
	return
}

func op_logical_and() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push_byte(b2y(y2b(a) && y2b(b)))
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
	iptr += s2i(offh, offl) - 1
	if iptr < 0 || iptr > len(script) {
		err = errors.New("jumped to invalid index")
	}
	return
}

func op_reg_store(reg byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	registers[reg] = a
	return
}

func op_reg_load(reg byte) (err error) {
	err = push(registers[reg])
	return
}

func op_reg_inc(reg, n byte) (err error) {
	registers[reg] = i2v(v2i(registers[reg]) + int64(n))
	return
}

func op_reg_dec(reg, n byte) (err error) {
	registers[reg] = i2v(v2i(registers[reg]) - int64(n))
	return
}

func op_data_move(loch, locl byte) (err error) {
	dptr += s2i(loch, locl)
	if dptr < 0 || dptr > len(script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_goto(loch, locl byte) (err error) {
	dptr = s2i(loch, locl)
	if dptr < 0 || dptr > len(script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_push(n byte) (err error) {
	var v value
	b := make([]byte, n)
	dptr += copy(b, script[dptr:])
	copy(v[:], b)
	err = push(v)
	return
}

func op_data_reg(n, reg byte) (err error) {
	var v value
	b := make([]byte, n)
	dptr += copy(b, script[dptr:])
	copy(v[:], b)
	registers[reg] = v
	return
}

func op_replace_byte() (err error) {
	err, a := op_pop()
	script[dptr] = a[0]
	return
}

func op_replace_short() (err error) {
	err, a := op_pop()
	script[dptr] = a[0]
	script[dptr+1] = a[1]
	return
}

func op_buf_copy(buf byte) (err error) {
	err, lengthv := op_pop()
	if err != nil {
		return
	}
	length := int16(v2i(lengthv))
	buffers[buf] = make([]byte, length)
	dptr += copy(buffers[buf], script[dptr:])
	return
}

func op_buf_paste(buf byte) (err error) {
	err, lengthv := op_pop()
	if err != nil {
		return
	}
	length := int(int16(v2i(lengthv)))
	// extend script if necessary
	if dptr+length > len(script) {
		ext := make([]byte, dptr+length-len(script))
		script = append(script, ext...)
	}
	b := make([]byte, length)
	copy(b, buffers[buf])
	copy(script[dptr:], b)
	return
}

func op_buf_prefix(buf byte) (err error) {
	err = op_data_push(0x02)
	if err != nil {
		return
	}
	err = op_buf_copy(buf)
	return
}

func op_buf_rest(buf byte) (err error) {
	buffers[buf] = make([]byte, len(script[dptr:]))
	dptr += copy(buffers[buf], script[dptr:])
	return
}

func op_transfer() (err error) {
	iptr = dptr - 1
	return
}

func op_reject() (err error) {
	return errors.New("rejectected input")
}

func op_add_sibling(buf byte) (err error) {
	encSibling := buffers[buf]
	print(len(encSibling))

	// decode sibling
	sib := new(quorum.Sibling)
	err = sib.GobDecode(encSibling)
	if err != nil {
		return
	}

	// add sibling
	q.AddSibling(wallet, sib)
	return
}

func op_add_wallet(buf byte) (err error) {
	_, lbalv := op_pop()
	_, ubalv := op_pop()
	err, idv := op_pop()
	if err != nil {
		return
	}

	// convert values to proper types
	id := quorum.WalletID(siaencoding.DecUint64(idv[:]))
	lbal := siaencoding.DecUint64(lbalv[:])
	ubal := siaencoding.DecUint64(ubalv[:])
	bal := quorum.NewBalance(ubal, lbal)

	// create wallet
	newscript := buffers[buf]
	_, err = q.CreateWallet(wallet, id, bal, newscript)
	return
}

func op_send() (err error) {
	_, lbalv := op_pop()
	_, ubalv := op_pop()
	err, idv := op_pop()
	if err != nil {
		return
	}

	// convert values to proper types
	id := quorum.WalletID(siaencoding.DecUint64(idv[:]))
	lbal := siaencoding.DecUint64(lbalv[:])
	ubal := siaencoding.DecUint64(ubalv[:])
	bal := quorum.NewBalance(ubal, lbal)

	// send
	_, err = q.Send(wallet, bal, id)
	return
}

func op_verify(pkey_buf, sm_buf byte) (err error) {
	// decode public key
	pk := new(siacrypto.PublicKey)
	err = pk.GobDecode(buffers[pkey_buf])
	if err != nil {
		return
	}
	// decode signed message
	// TODO: pack message and signature into one buffer?
	sm := new(siacrypto.SignedMessage)
	err = sm.GobDecode(buffers[sm_buf])
	if err != nil {
		return
	}
	// verify signature
	verified := pk.Verify(sm)
	err = op_push_byte(b2y(verified))
	return
}

func op_switch(cmp, offset byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	if cmp == a[0] {
		err = op_goto(0, offset)
	} else {
		err = push(a)
	}
	return
}
