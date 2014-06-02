package script

import (
	"errors"
	"reflect"
)

var opTable = []instruction{
	instruction{0x00, 0, reflect.ValueOf(op_nop), 1},
	instruction{0x01, 1, reflect.ValueOf(op_push), 2},
	instruction{0x02, 0, reflect.ValueOf(op_pop), 1},
	instruction{0x03, 0, reflect.ValueOf(op_dup), 1},
	instruction{0x04, 0, reflect.ValueOf(op_swap), 2},
	instruction{0x05, 0, reflect.ValueOf(op_add), 2},
	instruction{0x06, 0, reflect.ValueOf(op_sub), 2},
	instruction{0x07, 0, reflect.ValueOf(op_mul), 2},
	instruction{0x08, 0, reflect.ValueOf(op_div), 2},
	instruction{0x09, 0, reflect.ValueOf(op_mod), 3},
	instruction{0x0A, 0, reflect.ValueOf(op_neg), 1},
	instruction{0x0B, 0, reflect.ValueOf(op_eq), 2},
	instruction{0x0C, 0, reflect.ValueOf(op_ne), 2},
	instruction{0x0D, 0, reflect.ValueOf(op_lt), 2},
	instruction{0x0E, 0, reflect.ValueOf(op_gt), 2},
	instruction{0x0F, 0, reflect.ValueOf(op_not), 2},
	instruction{0x10, 0, reflect.ValueOf(op_or), 2},
	instruction{0x11, 0, reflect.ValueOf(op_and), 2},
	instruction{0x12, 1, reflect.ValueOf(op_if), 3},
	instruction{0x13, 0, reflect.ValueOf(op_goto), 2},
	instruction{0x14, 1, reflect.ValueOf(op_store), 2},
	instruction{0x15, 1, reflect.ValueOf(op_load), 2},
	instruction{0x16, 1, reflect.ValueOf(op_inc), 2},
	instruction{0x16, 1, reflect.ValueOf(op_dec), 2},
}

func op_nop() (err error) {
	return
}

func op_push(b byte) (err error) {
	stack = &stackElem{b, stack}
	stackLen++
	return
}

func op_pop() (err error, b byte) {
	if stackLen < 1 {
		err = errors.New("stack empty")
		return
	}
	b = stack.b
	stack = stack.next
	stackLen--
	return
}

func op_dup() (err error) {
	if stackLen < 1 {
		return errors.New("stack empty")
	}
	op_push(stack.b)
	return
}

func op_swap() (err error) {
	if stackLen < 2 {
		return errors.New("insufficient stack")
	}
	next := stack.next
	stack.next = stack.next.next
	next.next = stack
	stack = next
	return
}

func op_add() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(a + b)
	return
}

func op_sub() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(a - b)
	return
}

func op_mul() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(a * b)
	return
}

func op_div() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(a / b)
	return
}

func op_mod() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(a % b)
	return
}

// doesn't make a lot of sense for bytes...
func op_neg() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	op_push(-a)
	return
}

// helper function for booleans
func btoy(b bool) byte {
	if b {
		return 0x01
	} else {
		return 0x00
	}
}

func ytob(b byte) bool {
	return b != 0x00
}

func op_eq() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(a == b))
	return
}

func op_ne() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(a != b))
	return
}

func op_lt() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(a < b))
	return
}

func op_gt() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(a > b))
	return
}

func op_not() (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(!ytob(a)))
	return
}

func op_or() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(ytob(a) || ytob(b)))
	return
}

func op_and() (err error) {
	_, a := op_pop()
	err, b := op_pop()
	if err != nil {
		return
	}
	op_push(btoy(ytob(a) && ytob(b)))
	return
}

func op_if(offset byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	if ytob(a) {
		iptr += int(offset)
		if iptr < 0 {
			return errors.New("jumped to invalid index")
		}
	}
	return
}

func op_goto(offset byte) (err error) {
	iptr += int(offset)
	if iptr < 0 {
		return errors.New("jumped to invalid index")
	}
	return
}

func op_store(reg byte) (err error) {
	err, a := op_pop()
	if err != nil {
		return
	}
	registers[reg] = a
	return
}

func op_load(reg byte) (err error) {
	op_push(registers[reg])
	return
}

func op_inc(reg byte) (err error) {
	registers[reg]++
	return
}

func op_dec(reg byte) (err error) {
	registers[reg]--
	return
}
