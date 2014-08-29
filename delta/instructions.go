package delta

import (
	"bytes"
	"errors"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
)

var (
	errExit     = errors.New("exited")
	errRejected = errors.New("rejected input")
)

var opTable = map[byte]instruction{
	// general opcodes
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
	0x20: instruction{"if_move", 2, op_if_move, 2},
	0x21: instruction{"goto", 2, op_goto, 1},
	0x22: instruction{"move", 2, op_move, 1},
	0x23: instruction{"concat", 0, op_concat, 2},
	// data pointer and register opcodes
	0x30: instruction{"store", 1, op_store, 2},
	0x31: instruction{"load", 1, op_load, 2},
	0x32: instruction{"data_goto", 2, op_data_goto, 1},
	0x33: instruction{"data_move", 2, op_data_move, 1},
	0x34: instruction{"data_push", 1, op_data_push, 2},
	0x35: instruction{"data_store", 2, op_data_store, 2},
	0x36: instruction{"data_copy", 1, op_data_copy, 2},
	0x37: instruction{"data_paste", 1, op_data_paste, 2},
	0x38: instruction{"transfer", 0, op_transfer, 1},
	// function opcodes
	0x40: instruction{"verify", 0, op_verify, 9},
	0x41: instruction{"add_sibling", 0, op_add_sibling, 5},
	0x42: instruction{"add_wallet", 0, op_add_wallet, 5},
	0x43: instruction{"send", 0, op_send, 5},
	0x44: instruction{"update_sector", 0, op_update_sector, 9},
	0x46: instruction{"deadline", 0, op_deadline, 2},
	// convenience opcodes
	0xE0: instruction{"switch", 2, op_switch, 3},
	0xE1: instruction{"store_prefix", 1, op_store_prefix, 2},
	0xE2: instruction{"store_rest", 1, op_store_rest, 2},
	0xE3: instruction{"push_prefix", 0, op_push_prefix, 2},
	0xE4: instruction{"push_rest", 0, op_push_rest, 2},
	0xE5: instruction{"cond_reject", 0, op_cond_reject, 1},
	0xE6: instruction{"data_seek", 1, op_data_seek, 3},
	// termination opcodes
	0xFE: instruction{"reject", 0, op_reject, 0},
	0xFF: instruction{"exit", 0, op_exit, 0},
}

// helper functions

func v2i(b []byte) int64 {
	p := make([]byte, 8)
	copy(p, b)
	return siaencoding.DecInt64(p)
}

func i2v(i int64) []byte {
	return siaencoding.EncInt64(i)
}

func v2f(b []byte) float64 {
	p := make([]byte, 8)
	copy(p, b)
	return siaencoding.DecFloat64(p)
}

func f2v(f float64) []byte {
	return siaencoding.EncFloat64(f)
}

func s2i(low, high byte) int {
	return int(int16(low) + int16(high)<<8)
}

func b2v(b bool) []byte {
	if b {
		return []byte{0x01}
	}
	return []byte{0x00}
}

func v2b(b []byte) bool {
	return v2i(b) != 0
}

// general opcodes

func op_no_op(env *scriptEnv, args []byte) (err error) {
	return
}

func op_push_byte(env *scriptEnv, args []byte) (err error) {
	err = env.push(args[:1])
	return
}

func op_push_short(env *scriptEnv, args []byte) (err error) {
	err = env.push(args[:2])
	return
}

func op_pop(env *scriptEnv, args []byte) (err error) {
	_, err = env.pop()
	return
}

func op_dup(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	env.push(a)
	err = env.push(a)
	return
}

func op_swap(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(a)
	err = env.push(b)
	return
}

func op_add_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) + v2i(b)))
	return
}

func op_add_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(f2v(v2f(a) + v2f(b)))
	return
}

func op_sub_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) - v2i(b)))
	return
}

func op_sub_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(f2v(v2f(a) - v2f(b)))
	return
}

func op_mul_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) * v2i(b)))
	return
}

func op_mul_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(f2v(v2f(a) * v2f(b)))
	return
}

func op_div_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	if v2i(b) == 0 {
		return errors.New("divide by zero")
	}
	err = env.push(i2v(v2i(a) / v2i(b)))
	return
}

func op_div_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	if v2f(b) == 0.0 {
		return errors.New("divide by zero")
	}
	err = env.push(f2v(v2f(a) / v2f(b)))
	return
}

func op_mod_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) % v2i(b)))
	return
}

func op_neg_int(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(-v2i(a)))
	return
}

func op_neg_float(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(f2v(-v2f(a)))
	return
}

func op_binary_or(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) | v2i(b)))
	return
}

func op_binary_and(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) & v2i(b)))
	return
}

func op_binary_xor(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) ^ v2i(b)))
	return
}

func op_shift_left(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) << args[0]))
	return
}

func op_shift_right(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	err = env.push(i2v(v2i(a) >> args[0]))
	return
}

func op_equal(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(bytes.Equal(a, b)))
	return
}

func op_not_equal(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(!bytes.Equal(a, b)))
	return
}

func op_less_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2i(a) < v2i(b)))
	return
}

func op_less_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2f(a) < v2f(b)))
	return
}

func op_greater_int(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2i(a) > v2i(b)))
	return
}

func op_greater_float(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2f(a) > v2f(b)))
	return
}

func op_logical_not(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(!v2b(a)))
	return
}

func op_logical_or(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2b(a) || v2b(b)))
	return
}

func op_logical_and(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	env.push(b2v(v2b(a) && v2b(b)))
	return
}

func op_if_goto(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	if v2b(a) {
		err = op_goto(env, args)
	}
	return
}

func op_if_move(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	if v2b(a) {
		err = op_move(env, args)
	}
	return
}

func op_goto(env *scriptEnv, args []byte) (err error) {
	env.iptr = s2i(args[0], args[1]) - 1
	if env.iptr < 0 || env.iptr > len(env.script) {
		err = errors.New("jumped to invalid index")
	}
	return
}

func op_move(env *scriptEnv, args []byte) (err error) {
	env.iptr += s2i(args[0], args[1]) - 1
	if env.iptr < 0 || env.iptr > len(env.script) {
		err = errors.New("jumped to invalid index")
	}
	return
}

func op_concat(env *scriptEnv, args []byte) (err error) {
	a, _ := env.pop()
	b, err := env.pop()
	if err != nil {
		return
	}
	return env.push(append(b, a...))
}

// data pointer and register opcodes

func op_store(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	env.registers[args[0]] = a
	return
}

func op_load(env *scriptEnv, args []byte) (err error) {
	err = env.push(env.registers[args[0]])
	return
}

func op_data_goto(env *scriptEnv, args []byte) (err error) {
	env.dptr = s2i(args[0], args[1])
	if env.dptr < 0 || env.dptr > len(env.script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_move(env *scriptEnv, args []byte) (err error) {
	env.dptr += s2i(args[0], args[1])
	if env.dptr < 0 || env.dptr > len(env.script) {
		err = errors.New("invalid data access")
	}
	return
}

func op_data_push(env *scriptEnv, args []byte) (err error) {
	b := make([]byte, args[0])
	env.dptr += copy(b, env.script[env.dptr:])
	err = env.push(b)
	return
}

func op_data_store(env *scriptEnv, args []byte) (err error) {
	b := make([]byte, args[0])
	env.dptr += copy(b, env.script[env.dptr:])
	env.registers[args[1]] = b
	return
}

func op_data_copy(env *scriptEnv, args []byte) (err error) {
	lengthv, err := env.pop()
	if err != nil {
		return
	}
	length := uint16(v2i(lengthv))
	env.registers[args[0]] = make([]byte, length)
	env.dptr += copy(env.registers[args[0]], env.script[env.dptr:])
	return
}

func op_data_paste(env *scriptEnv, args []byte) (err error) {
	lengthv, err := env.pop()
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
	copy(b, env.registers[args[0]])
	copy(env.script[env.dptr:], b)
	return
}

func op_transfer(env *scriptEnv, args []byte) (err error) {
	env.iptr = env.dptr
	return
}

// function opcodes

func op_add_sibling(env *scriptEnv, args []byte) (err error) {
	encSib, err := env.pop()
	if err != nil {
		return
	}

	var sib state.Sibling
	err = siaencoding.Unmarshal(encSib, &sib)
	if err != nil {
		return
	}

	env.engine.AddSibling(env.wallet, sib)
	return
}

func op_add_wallet(env *scriptEnv, args []byte) (err error) {
	// pop values
	script, _ := env.pop()
	balb, _ := env.pop()
	idb, err := env.pop()
	if err != nil {
		return
	}

	// convert values to proper types
	var bal state.Balance
	copy(bal[:], balb)
	encUint64 := make([]byte, 8)
	copy(encUint64, idb)
	id := state.WalletID(siaencoding.DecUint64(encUint64))

	// call API function
	return env.engine.CreateWallet(env.wallet, id, bal, script)
}

func op_send(env *scriptEnv, args []byte) (err error) {
	balb, _ := env.pop()
	idb, err := env.pop()
	if err != nil {
		return
	}

	var bal state.Balance
	copy(bal[:], balb)
	id := state.WalletID(siaencoding.DecUint64(idb))

	err = env.engine.SendCoin(env.wallet, bal, id)
	return
}

func op_verify(env *scriptEnv, args []byte) (err error) {
	msg, _ := env.pop()
	sigBytes, _ := env.pop()
	pkBytes, err := env.pop()
	if err != nil {
		return
	}

	var pk siacrypto.PublicKey
	copy(pk[:], pkBytes)
	var sig siacrypto.Signature
	copy(sig[:], sigBytes)

	verified := pk.Verify(sig, msg)

	// push success value
	err = env.push(b2v(verified))
	return
}

func op_update_sector(env *scriptEnv, args []byte) (err error) {
	deadline, _ := env.pop()
	confreq, _ := env.pop()
	hashset, _ := env.pop()
	d, _ := env.pop()
	k, _ := env.pop()
	atoms, err := env.pop()
	if err != nil {
		return
	}

	if len(atoms) != 2 ||
		len(k) != 1 || len(d) != 1 ||
		len(hashset) != int(state.QuorumSize)*siacrypto.HashSize ||
		len(confreq) != 1 ||
		len(deadline) != 4 {
		err = errors.New("invalid parameter")
		return
	}

	var hs [state.QuorumSize]siacrypto.Hash
	for i := range hs {
		copy(hs[i][:], hashset)
		hashset = hashset[siacrypto.HashSize:]
	}

	su := state.SectorUpdate{
		Atoms:                 siaencoding.DecUint16(atoms),
		K:                     k[0],
		D:                     d[0],
		HashSet:               hs,
		ConfirmationsRequired: confreq[0],
	}
	su.Event.Deadline = siaencoding.DecUint32(deadline)

	err = env.engine.UpdateSector(env.wallet, su)
	return
}

func op_deadline(env *scriptEnv, args []byte) (err error) {
	return env.push(siaencoding.EncUint32(env.deadline))
}

// convenience opcodes

func op_switch(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	if args[0] == a[0] {
		err = op_goto(env, []byte{0, args[1]})
	} else {
		err = env.push(a)
	}
	return
}

func op_store_prefix(env *scriptEnv, args []byte) (err error) {
	err = op_data_push(env, []byte{0x02})
	if err != nil {
		return
	}
	err = op_data_copy(env, args)
	return
}

func op_store_rest(env *scriptEnv, args []byte) (err error) {
	env.registers[args[0]] = make([]byte, len(env.script[env.dptr:]))
	copy(env.registers[args[0]], env.script[env.dptr:])
	return
}

func op_push_prefix(env *scriptEnv, args []byte) (err error) {
	num := s2i(env.script[env.dptr], env.script[env.dptr+1])
	env.dptr += 2
	b := make([]byte, num)
	env.dptr += copy(b, env.script[env.dptr:])
	err = env.push(b)
	return
}

func op_push_rest(env *scriptEnv, args []byte) (err error) {
	b := make([]byte, len(env.script[env.dptr:]))
	copy(b, env.script[env.dptr:])
	err = env.push(b)
	return
}

func op_cond_reject(env *scriptEnv, args []byte) (err error) {
	a, err := env.pop()
	if err != nil {
		return
	}
	if !v2b(a) {
		err = op_reject(env, []byte{})
	}
	return
}

func op_data_seek(env *scriptEnv, args []byte) (err error) {
	i := bytes.IndexByte(env.script[env.dptr:], args[0])
	if i == -1 {
		return errors.New("no marker found")
	}
	env.dptr += i
	// advance past marker byte itself, since we usually don't want to read it
	env.dptr++

	// skip the occurrence in the opcode itself
	if env.script[env.dptr-2] == 0xE6 {
		return op_data_seek(env, args)
	}
	return
}

// termination opcodes

func op_reject(env *scriptEnv, args []byte) (err error) {
	return errRejected
}

func op_exit(env *scriptEnv, args []byte) (err error) {
	return errExit
}
