package script

import (
	"errors"
	"quorum"
	"reflect"
	"runtime" // for debugging
)

const (
	MaxInstructions = 10000
	MaxStackLen     = 1 << 16
)

type Script struct {
	Block []byte
}

type ScriptInput struct {
	WalletID quorum.WalletID
	Input    []byte
}

type instruction struct {
	opcode   byte
	argBytes int
	fn       reflect.Value
	cost     int
}

func (i *instruction) print(args []reflect.Value) {
	fnName := runtime.FuncForPC(i.fn.Pointer()).Name()
	print(fnName[14:])
	for i := range args {
		print(" ", args[i].Uint())
	}
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

func (s *stackElem) print() {
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
// TODO: replace with env struct
var (
	script    []byte
	iptr      int
	dptr      int
	registers [256]value
	buffer    []byte
	stack     *stackElem
	stackLen  int
	q         *quorum.Quorum
	// resource pools
	instBalance int
	costBalance int
)

// deduct instruction cost from resource pools, and return an error if any pool is exhausted
func deductResources(op instruction) error {
	instBalance -= 1
	costBalance -= op.cost
	switch {
	case instBalance < 0:
		return errors.New("instruction limit reached")
	case costBalance < 0:
		return errors.New("balance exhausted")
	default:
		return nil
	}
}

// Execute interprets a script on a set of inputs and returns the execution cost.
func (s *Script) Execute(in []byte, q_ *quorum.Quorum) (totalCost int, err error) {
	// initialize execution environment
	script = append(s.Block, in...)
	iptr = 0
	dptr = len(s.Block)
	registers = [256]value{}
	buffer = nil
	stack = nil
	stackLen = 0
	q = q_
	// resource pools
	// these values will likely be supplied as arguments in the future
	instBalance = MaxInstructions
	costBalance = 10000

	for {
		if iptr >= len(script) {
			err = errors.New("script missing terminator")
			break
		} else if script[iptr] == 0xFF {
			break
		}

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

		// deduct resources and check that we can proceed with execution
		err = deductResources(op)
		if err != nil {
			break
		}

		// call associated opcode function
		retVals := op.fn.Call(fnArgs)
		errInter := retVals[0].Interface()
		if errInter != nil {
			err = errInter.(error)
			break
		}

		// DEBUG: print op and stack
		//op.print(fnArgs)
		//print("\n    -> ")
		//stack.print()

		// increment instruction pointer
		iptr++
	}

	return
}
