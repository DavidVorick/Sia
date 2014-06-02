package script

import (
	"errors"
	"reflect"
)

const MaxInstructions = 10000

type instruction struct {
	opcode   byte
	argBytes int
	fn       reflect.Value
	cost     int
}

type stackElem struct {
	b    byte
	next *stackElem
}

func (s *stackElem) Print() {
	print("{ ")
	p := s
	for {
		if p == nil {
			break
		}
		print(p.b, " ")
		p = p.next
	}
	println("}")
}

// global vars accessed by the various opcode functions
var (
	stack     *stackElem
	stackLen  int
	iptr      int
	opCounts  map[byte]int
	registers [256]byte
)

func ExecuteScript(script []byte) (totalCost int, err error) {
	// initialize execution environment
	stack = nil
	stackLen = 0
	iptr = 0
	opCounts = make(map[byte]int)
	registers = [256]byte{}

	for {
		if iptr > len(script) {
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

		// place arguments in array, advance instruction pointer
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
		//stack.Print()
	}

	// calculate cost
	for b, n := range opCounts {
		totalCost += n * opTable[b].cost
	}
	return
}
