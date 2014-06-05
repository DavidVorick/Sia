package script

import (
	"errors"
	"quorum"
	"reflect"
)

const (
	MaxInstructions = 10000
	MaxStackLen     = 1 << 16
)

type Script struct {
	// Wallet quorum.WalletID
	Block []byte
}

type instruction struct {
	opcode   byte
	argBytes int
	fn       reflect.Value
	cost     int
}

// generic 64-bit value
type value [8]byte

type stackElem struct {
	val  value
	next *stackElem
}

func push(b value) (err error) {
	if stackLen > MaxStackLen {
		return errors.New("stack overflow")
	}
	stack = &stackElem{b, stack}
	stackLen++
	return
}

func (s *stackElem) Print() {
	print("{ ")
	p := s
	for {
		if p == nil {
			break
		}
		print(v2i(p.val), " ")
		p = p.next
	}
	println("}")
}

// global vars accessed by the various opcode functions
var (
	script    []byte
	iptr      int
	input     []byte
	registers [256]value
	stack     *stackElem
	stackLen  int
	q         *quorum.Quorum
	opCounts  map[byte]int
)

func (s *Script) Bytes() []byte {
	return s.Block
}

func (s *Script) Execute(in []byte, quorum *quorum.Quorum) (totalCost int, err error) {
	// initialize execution environment
	script = s.Block
	iptr = 0
	input = in
	registers = [256]value{}
	stack = nil
	stackLen = 0
	q = quorum
	opCounts = make(map[byte]int)

	for {
		if iptr >= len(script) {
			err = errors.New("script missing terminator")
			break
		} else if iptr > MaxInstructions {
			err = errors.New("max instruction limit reached")
			break
		} else if script[iptr] == 0xFF {
			break
		}

		// record in map
		opCounts[script[iptr]]++

		op := opTable[script[iptr]]

		// place arguments in array while advancing instruction pointer
		if iptr+op.argBytes >= len(script) {
			err = errors.New("too few arguments to opcode")
			break
		}
		var fnArgs []reflect.Value
		for j := 0; j < op.argBytes; j++ {
			iptr++
			fnArgs = append(fnArgs, reflect.ValueOf(script[iptr]))
		}

		// call associated opcode function
		retVals := op.fn.Call(fnArgs)
		errInter := retVals[0].Interface()
		if errInter != nil {
			err = errInter.(error)
			return
		}

		// increment instruction pointer
		iptr++

		// DEBUG: print stack
		stack.Print()
	}

	// calculate cost
	for b, n := range opCounts {
		totalCost += n * opTable[b].cost
	}
	return
}
