package script

import (
	"errors"
	"fmt"
	"quorum"
	"reflect"
)

const (
	MaxInstructions = 10000
	MaxStackLen     = 1 << 16
	DEBUG           = false
)

type ScriptInput struct {
	WalletID quorum.WalletID
	Input    []byte
}

type instruction struct {
	opcode   byte
	name     string
	argBytes int
	fn       reflect.Value
	cost     int
}

func (in *instruction) print(args []reflect.Value) string {
	s := in.name
	for i := range args {
		s += fmt.Sprint(" ", args[i].Uint())
	}
	return s
}

// generic 64-bit value
type value [8]byte

type stackElem struct {
	val  value
	next *stackElem
}

func push(v value) (err error) {
	if stackLen > MaxStackLen {
		return errors.New("stack overflow")
	}
	stack = &stackElem{v, stack}
	stackLen++
	return
}

func (s *stackElem) print() string {
	str := "{ "
	p := s
	for {
		if p == nil {
			break
		}
		str += fmt.Sprint(v2i(p.val))
		str += " "
		p = p.next
	}
	str += "}"
	return str
}

// global vars accessed by the various opcode functions
// TODO: replace with env struct
var (
	script    []byte
	iptr      int
	dptr      int
	registers [256]value
	buffers   [256][]byte
	stack     *stackElem
	stackLen  int
	wallet    *quorum.Wallet
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
func (si *ScriptInput) Execute(q_ *quorum.Quorum) (totalCost int, err error) {
	if si == nil {
		err = errors.New("nil ScriptInput")
	}
	// initialize execution environment
	q = q_
	wallet = q.LoadWallet(si.WalletID)
	if wallet == nil {
		err = errors.New("failed to load wallet")
		return
	}
	script = append(wallet.Script(), si.Input...)
	iptr = 0
	dptr = len(wallet.Script())
	registers = [256]value{}
	buffers = [256][]byte{}
	stack = nil
	stackLen = 0
	// resource pools
	// these values will likely be supplied as arguments in the future
	instBalance = MaxInstructions
	costBalance = 10000
	fmt.Println("executing script:", script)

	for {
		if iptr >= len(script) || script[iptr] == 0xFF {
			break
		}

		if int(script[iptr]) > len(opTable) {
			err = errors.New("invalid opcode " + fmt.Sprint(script[iptr]))
			break
		}
		op := opTable[script[iptr]]

		// place arguments in array while advancing instruction pointer
		if iptr+op.argBytes >= len(script) {
			err = errors.New("too few arguments to opcode " + op.name)
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
			err = errors.New("instruction \"" + op.print(fnArgs) + "\" failed: " + errInter.(error).Error())
			break
		}

		if DEBUG {
			fmt.Println(op.print(fnArgs))
			fmt.Println("    stack:", stack.print())
			b := make([]byte, 20)
			copy(b, buffers[1])
			fmt.Println("    buffer 1:", len(buffers[1]), b)
			b = make([]byte, 20)
			copy(b, buffers[2])
			fmt.Println("    buffer 2:", len(buffers[2]), b)
		}
		// increment instruction pointer
		iptr++
	}
	if err != nil {
		fmt.Println("script execution failed:", err)
	}
	q.SaveWallet(wallet)
	return
}
