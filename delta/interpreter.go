package delta

import (
	"errors"
	"fmt"

	"github.com/NebulousLabs/Sia/state"
)

const (
	maxInstructions = 10000
	maxStackLen     = 1 << 16
	debug           = true
)

// A ScriptInput pairs an input byte slice with the WalletID associated with
// the recipient. During execution, the WalletID is used to load the script
// body, and then the Input is appended to the end of the script.
type ScriptInput struct {
	WalletID state.WalletID
	Input    []byte
}

type instruction struct {
	name     string
	argBytes int
	fn       func([]byte) error
	cost     int
}

func (in *instruction) print(args []byte) string {
	s := in.name
	for _, b := range args {
		s += fmt.Sprint(" ", b)
	}
	return s
}

type stackElem struct {
	val  []byte
	next *stackElem
}

func push(v []byte) (err error) {
	if env.stackLen > maxStackLen {
		return errors.New("stack overflow")
	}
	c := make([]byte, len(v))
	copy(c, v)
	env.stack = &stackElem{c, env.stack}
	env.stackLen++
	return
}

func pop() (v []byte, err error) {
	if env.stackLen < 1 {
		err = errors.New("stack empty")
		return
	}
	v = env.stack.val
	env.stack = env.stack.next
	env.stackLen--
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

// environment variables necessary for script execution
type scriptEnv struct {
	script     []byte
	iptr, dptr int
	registers  [256][]byte
	stack      *stackElem
	stackLen   int
	wallet     state.Wallet
	state      *state.State
	// resource pools
	instBalance int
	costBalance int
}

// global execution environment
var env scriptEnv

// deduct instruction cost from resource pools, and return an error if any pool
// is exhausted.
// TODO: add memBalance to prevent buffer abuse
func deductResources(op instruction) error {
	env.instBalance--
	env.costBalance -= op.cost
	switch {
	case env.instBalance < 0:
		return errors.New("instruction limit reached")
	case env.costBalance < 0:
		return errors.New("balance exhausted")
	default:
		return nil
	}
}

// Execute loads the requested script, appends the script input data, sets up
// an execution environment, and interprets bytecodes until a termination
// condition is reached.
func (e *Engine) Execute(si ScriptInput) (totalCost int, err error) {
	// load wallet
	w, err := e.state.LoadWallet(si.WalletID)
	if err != nil {
		return
	}

	// initialize execution environment
	env = scriptEnv{
		script: append(w.Script, si.Input...),
		dptr:   len(w.Script),
		wallet: w,
		state:  &e.state,
		// these values will likely be stored as part of the wallet
		instBalance: maxInstructions,
		costBalance: 10000,
	}

	// run script
	fmt.Println("executing script:", env.script)
	if err = env.run(); err != nil {
		fmt.Println("script execution failed:", err)
	}

	e.state.SaveWallet(env.wallet)
	return
}

// run performs the actual execution of opcodes.
func (env *scriptEnv) run() error {
	for {
		if env.iptr >= len(env.script) {
			return nil
		}

		// look up opcode
		op, ok := opTable[env.script[env.iptr]]
		if !ok {
			return errors.New("invalid opcode " + fmt.Sprint(env.script[env.iptr]))
		}

		if env.iptr+op.argBytes >= len(env.script) {
			return errors.New("too few arguments to opcode " + op.name)
		}

		// deduct resources and check that we can proceed with execution
		if err := deductResources(op); err != nil {
			return err
		}

		// read bytes into argument array and advance env.iptr
		fnArgs := make([]byte, op.argBytes)
		env.iptr++
		env.iptr += copy(fnArgs, env.script[env.iptr:])

		// call associated opcode function and check for error
		if err := op.fn(fnArgs); err != nil {
			switch err {
			case errExit:
				return nil
			case errRejected:
				return err
			default:
				return errors.New("instruction \"" + op.print(fnArgs) + "\" failed: " + err.Error())
			}
		}

		if debug {
			fmt.Println(op.print(fnArgs))
			fmt.Println("    stack:", env.stack.print())
			b := make([]byte, 20)
			copy(b, env.registers[1])
			fmt.Println("    register 1:", len(env.registers[1]), b)
			b = make([]byte, 20)
			copy(b, env.registers[2])
			fmt.Println("    register 2:", len(env.registers[2]), b)
		}
	}
}
