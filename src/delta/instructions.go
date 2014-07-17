package delta

import (
	"errors"
	"siacrypto"
	"siaencoding"
)

var (
	errExit     = errors.New("exited")
	errRejected = errors.New("rejected input")
)

var opTable = map[byte]instruction{
	0x00: instruction{"no_op", 0, op_no_op, 1},
	0x01: instruction{"push_byte", 1, op_push_byte, 2},
	0x02: instruction{"push_short", 2, op_push_short, 2},
	0x03: instruction{"pop", 0, op_pop, 1},
	0x04: instruction{"dup", 0, op_dup, 2},
	0x05: instruction{"swap", 0, op_swap, 2},
	0x06: instruction{"add_int", 0, op_add_int, 2},
	0x07: instruction{"add_float", 0, op_add_float, 3},
	0x08: instruction{"sub_int", 0, op_sub_int, 2},
	0x09: instruction{"sub_float", 0, op_sub_float, 3},
	0x0A: instruction{"mul_int", 0, op_mul_int, 2},
	0x0B: instruction{"mul_float", 0, op_mul_float, 3},
	0x0C: instruction{"div_int", 0, op_div_int, 2},
	0x0D: instruction{"div_float", 0, op_div_float, 3},
	0x0E: instruction{"mod_int", 0, op_mod_int, 3},
	0x0F: instruction{"neg_int", 0, op_neg_int, 2},
	0x10: instruction{"neg_float", 0, op_neg_float, 3},
	0x11: instruction{"binary_or", 0, op_binary_or, 2},
	0x12: instruction{"binary_and", 0, op_binary_and, 2},
	0x13: instruction{"binary_xor", 0, op_binary_xor, 2},
	0x14: instruction{"shift_left", 1, op_shift_left, 2},
	0x15: instruction{"shift_right", 1, op_shift_right, 2},
	0x16: instruction{"equal", 0, op_equal, 2},
	0x17: instruction{"not_equal", 0, op_not_equal, 2},
	0x18: instruction{"less_int", 0, op_less_int, 2},
	0x19: instruction{"less_float", 0, op_less_float, 2},
	0x1A: instruction{"greater_int", 0, op_greater_int, 2},
	0x1B: instruction{"greater_float", 0, op_greater_float, 2},
	0x1C: instruction{"logical_not", 0, op_logical_not, 2},
	0x1D: instruction{"logical_or", 0, op_logical_or, 2},
	0x1E: instruction{"logical_and", 0, op_logical_and, 2},
	0x1F: instruction{"if_goto", 2, op_if_goto, 2},
	0x20: instruction{"goto", 2, op_goto, 1},
	0x21: instruction{"reg_store", 1, op_reg_store, 2},
	0x22: instruction{"reg_load", 1, op_reg_load, 2},
	0x23: instruction{"reg_inc", 1, op_reg_inc, 2},
	0x24: instruction{"reg_dec", 1, op_reg_dec, 2},
	0x25: instruction{"data_move", 2, op_data_move, 1},
	0x26: instruction{"data_goto", 2, op_data_goto, 1},
	0x27: instruction{"data_push", 1, op_data_push, 2},
	0x28: instruction{"data_reg", 2, op_data_reg, 2},
	0x29: instruction{"replace_byte", 0, op_replace_byte, 2},
	0x2A: instruction{"replace_short", 0, op_replace_short, 2},
	0x2B: instruction{"buf_copy", 1, op_buf_copy, 2},
	0x2C: instruction{"buf_paste", 1, op_buf_paste, 2},
	0x2D: instruction{"buf_prefix", 1, op_buf_prefix, 2},
	0x2E: instruction{"buf_rest", 1, op_buf_rest, 2},
	0x2F: instruction{"transfer", 0, op_transfer, 1},
	0x30: instruction{"reject", 0, op_reject, 0},
	0x31: instruction{"add_sibling", 1, op_add_sibling, 5},
	0x32: instruction{"add_wallet", 1, op_add_wallet, 5},
	0x33: instruction{"send", 0, op_send, 5},
	0x34: instruction{"verify", 2, op_verify, 9},
	0x35: instruction{"switch", 2, op_switch, 3},
	0x36: instruction{"if_move", 2, op_if_move, 2},
	0x37: instruction{"move", 2, op_move, 1},
	0x38: instruction{"cond_reject", 0, op_cond_reject, 1},
	0x39: instruction{"data_buf", 2, op_data_buf, 2},
	0x3A: instruction{"resize_sec", 1, op_resize_sec, 9},
	0x3B: instruction{"prop_upload", 1, op_prop_upload, 9},
	0xFF: instruction{"exit", 0, op_exit, 0},
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
func s2i(low, high byte) int {
	return int(int16(low) + int16(high)<<8)
}

func b2v(b bool) value {
	if b {
		return value{0x01}
	} else {
		return value{0x00}
	}
}

func v2b(v value) bool {
	return v2i(v) != 0
}

// opcodes

func op_no_op(args []byte) (err error) {
	return
}

func op_push_byte(args []byte) (err error) {
	err = push(value{args[0]})
	return
}

func op_push_short(args []byte) (err error) {
	err = push(value{args[0], args[1]})
	return
}

func op_pop(args []byte) (err error) {
	_, err = pop()
	return
}

func op_dup(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	push(a)
	err = push(a)
	return
}

func op_swap(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(a)
	err = push(b)
	return
}

func op_add_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) + v2i(b)))
	return
}

func op_add_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) + v2f(b)))
	return
}

func op_sub_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) - v2i(b)))
	return
}

func op_sub_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) - v2f(b)))
	return
}

func op_mul_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) * v2i(b)))
	return
}

func op_mul_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(f2v(v2f(a) * v2f(b)))
	return
}

func op_div_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	if v2i(b) == 0 {
		return errors.New("divide by zero")
	}
	err = push(i2v(v2i(a) / v2i(b)))
	return
}

func op_div_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	if v2f(b) == 0.0 {
		return errors.New("divide by zero")
	}
	err = push(f2v(v2f(a) / v2f(b)))
	return
}

func op_mod_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) % v2i(b)))
	return
}

func op_neg_int(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(-v2i(a)))
	return
}

func op_neg_float(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	err = push(f2v(-v2f(a)))
	return
}

func op_binary_or(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) | v2i(b)))
	return
}

func op_binary_and(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) & v2i(b)))
	return
}

func op_binary_xor(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) ^ v2i(b)))
	return
}

func op_shift_left(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) << args[0]))
	return
}

func op_shift_right(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	err = push(i2v(v2i(a) >> args[0]))
	return
}

func op_equal(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(a == b))
	return
}

func op_not_equal(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(a != b))
	return
}

func op_less_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2i(a) < v2i(b)))
	return
}

func op_less_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2f(a) < v2f(b)))
	return
}

func op_greater_int(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2i(a) > v2i(b)))
	return
}

func op_greater_float(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2f(a) > v2f(b)))
	return
}

func op_logical_not(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	push(b2v(!v2b(a)))
	return
}

func op_logical_or(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2b(a) || v2b(b)))
	return
}

func op_logical_and(args []byte) (err error) {
	a, _ := pop()
	b, err := pop()
	if err != nil {
		return
	}
	push(b2v(v2b(a) && v2b(b)))
	return
}

func op_if_goto(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	if v2b(a) {
		err = op_goto(args)
	}
	return
}

func op_goto(args []byte) (err error) {
	env.iptr = s2i(args[0], args[1]) - 1
	if env.iptr < 0 || env.iptr > len(env.script) {
		err = errors.New("jumped to invalid index")
	}
	return
}

func op_if_move(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	if v2b(a) {
		err = op_move(args)
	}
	return
}

func op_move(args []byte) (err error) {
	env.iptr += s2i(args[0], args[1]) - 1
	if env.iptr < 0 || env.iptr > len(env.script) {
		err = errors.New("jumped to invalid index")
	}
	return
}

func op_reg_store(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	env.registers[args[0]] = a
	return
}

func op_reg_load(args []byte) (err error) {
	err = push(env.registers[args[0]])
	return
}

func op_reg_inc(args []byte) (err error) {
	env.registers[args[0]] = i2v(v2i(env.registers[args[0]]) + int64(args[1]))
	return
}

func op_reg_dec(args []byte) (err error) {
	env.registers[args[0]] = i2v(v2i(env.registers[args[0]]) - int64(args[1]))
	return
}

func op_data_move(args []byte) (err error) {
	env.dptr += s2i(args[0], args[1])
	if env.dptr < 0 || env.dptr > len(env.script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_goto(args []byte) (err error) {
	env.dptr = s2i(args[0], args[1])
	if env.dptr < 0 || env.dptr > len(env.script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_push(args []byte) (err error) {
	var v value
	b := make([]byte, args[0])
	env.dptr += copy(b, env.script[env.dptr:])
	copy(v[:], b)
	err = push(v)
	return
}

func op_data_reg(args []byte) (err error) {
	var v value
	b := make([]byte, args[0])
	env.dptr += copy(b, env.script[env.dptr:])
	copy(v[:], b)
	env.registers[args[1]] = v
	return
}

func op_replace_byte(args []byte) (err error) {
	a, err := pop()
	env.script[env.dptr] = a[0]
	return
}

func op_replace_short(args []byte) (err error) {
	a, err := pop()
	env.script[env.dptr] = a[0]
	env.script[env.dptr+1] = a[1]
	return
}

func op_data_buf(args []byte) (err error) {
	env.buffers[args[1]] = make([]byte, args[0])
	env.dptr += copy(env.buffers[args[1]], env.script[env.dptr:])
	return
}

func op_buf_copy(args []byte) (err error) {
	lengthv, err := pop()
	if err != nil {
		return
	}
	length := uint16(v2i(lengthv))
	env.buffers[args[0]] = make([]byte, length)
	env.dptr += copy(env.buffers[args[0]], env.script[env.dptr:])
	return
}

func op_buf_paste(args []byte) (err error) {
	lengthv, err := pop()
	if err != nil {
		return
	}
	length := int(int16(v2i(lengthv)))
	// extend env.script if necessary
	if env.dptr+length > len(env.script) {
		ext := make([]byte, env.dptr+length-len(env.script))
		env.script = append(env.script, ext...)
	}
	b := make([]byte, length)
	copy(b, env.buffers[args[0]])
	copy(env.script[env.dptr:], b)
	return
}

func op_buf_prefix(args []byte) (err error) {
	err = op_data_push([]byte{0x02})
	if err != nil {
		return
	}
	err = op_buf_copy(args)
	return
}

func op_buf_rest(args []byte) (err error) {
	env.buffers[args[0]] = make([]byte, len(env.script[env.dptr:]))
	env.dptr += copy(env.buffers[args[0]], env.script[env.dptr:])
	return
}

func op_transfer(args []byte) (err error) {
	env.iptr = env.dptr
	return
}

func op_reject(args []byte) (err error) {
	return errRejected
}

func op_cond_reject(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	if !v2b(a) {
		err = op_reject([]byte{})
	}
	return
}

func op_add_sibling(args []byte) (err error) {
	/*
		// decode sibling
		sib := new(quorum.Sibling)
		err = sib.GobDecode(env.buffers[args[0]])
		if err != nil {
			return
		}

		// add sibling
		env.quorum.AddSibling(env.wallet, sib)
	*/
	return
}

func op_add_wallet(args []byte) (err error) {
	/*
		lbalv, _ := pop()
		ubalv, _ := pop()
		idv, err := pop()
		if err != nil {
			return
		}

		// convert values to proper types
		id := quorum.WalletID(siaencoding.DecUint64(idv[:]))
		lbal := siaencoding.DecUint64(lbalv[:])
		ubal := siaencoding.DecUint64(ubalv[:])
		bal := quorum.NewBalance(ubal, lbal)

		// create env.wallet
		newscript := env.buffers[args[0]]
		_, err = env.quorum.CreateWallet(env.wallet, id, bal, newscript)
	*/
	return
}

func op_send(args []byte) (err error) {
	/*
		lbalv, _ := pop()
		ubalv, _ := pop()
		idv, err := pop()
		if err != nil {
			return
		}

		// convert values to proper types
		id := quorum.WalletID(siaencoding.DecUint64(idv[:]))
		lbal := siaencoding.DecUint64(lbalv[:])
		ubal := siaencoding.DecUint64(ubalv[:])
		bal := quorum.NewBalance(ubal, lbal)

		// send
		_, err = env.quorum.Send(env.wallet, bal, id)
	*/
	return
}

func op_verify(args []byte) (err error) {
	// get public key
	var pk siacrypto.PublicKey
	copy(pk[:], env.buffers[args[0]])
	// decode signed message
	var sm siacrypto.SignedMessage
	err = siaencoding.Unmarshal(env.buffers[args[1]], &sm)
	if err != nil {
		return
	}
	// verify signature
	verified := pk.Verify(sm)
	err = push(b2v(verified))
	return
}

func op_switch(args []byte) (err error) {
	a, err := pop()
	if err != nil {
		return
	}
	if args[0] == a[0] {
		err = op_goto([]byte{0, args[1]})
	} else {
		err = push(a)
	}
	return
}

func op_resize_sec(args []byte) (err error) {
	/*
		a, err := pop()
		if err != nil {
			return
		}
		atoms := siaencoding.DecUint16(a[:2])
		_, _, err = env.quorum.ResizeSectorErase(env.wallet, atoms, args[0])
	*/
	return
}

func op_prop_upload(args []byte) (err error) {
	/*
		// decode function arguments
		var ua quorum.UploadArgs
		err = ua.GobDecode(env.buffers[args[0]])
		if err != nil {
			return
		}

		// call function
		_, _, err = env.quorum.ProposeUpload(
			env.wallet,
			ua.ParentHash,
			ua.NewHashSet,
			ua.AtomsChanged,
			ua.Confirmations,
			ua.Deadline,
		)
	*/
	return
}

func op_exit(b []byte) (err error) {
	return errExit
}
