package delta

import (
	"errors"
	"fmt"

	"github.com/NebulousLabs/Sia/state"
)

const (
	maxInstructions = 10000
	maxMemory       = 1 << 14 // 16 KB
	maxStackLen     = 1 << 16
	debug           = false
)

type instruction struct {
	name     string
	argBytes int
	fn       func(*scriptEnv, []byte) error
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

func (env *scriptEnv) push(v []byte) (err error) {
	if env.stackLen > maxStackLen {
		return errors.New("stack overflow")
	}
	c := make([]byte, len(v))
	copy(c, v)
	env.stack = &stackElem{c, env.stack}
	env.stackLen++
	return
}

func (env *scriptEnv) pop() (v []byte, err error) {
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
	for p := s; p != nil; p = p.next {
		if p == nil {
			break
		}
		if len(p.val) > 5 {
			str += fmt.Sprint(p.val[:5]) + "... "
		} else {
			str += fmt.Sprint(p.val) + " "
		}
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
	wallet     *state.Wallet
	engine     *Engine
	deadline   uint32
	// resource pools
	instBalance int
	costBalance int
	memUsage    int
}

// deduct instruction cost from resource pools, and return an error if any pool
// is exhausted.
func (env *scriptEnv) deductResources(op instruction) error {
	env.instBalance--
	env.costBalance -= op.cost
	// calculate memUsage
	env.memUsage = 0
	for i := range env.registers {
		env.memUsage += len(env.registers[i])
	}
	p := env.stack
	for {
		if p == nil {
			break
		}
		env.memUsage += len(p.val)
		p = p.next
	}

	// check each resource for exhaustion
	// TODO: export these errors
	switch {
	case env.instBalance < 0:
		return errors.New("instruction limit reached")
	case env.costBalance < 0:
		return errors.New("balance exhausted")
	case env.memUsage > maxMemory:
		return errors.New("memory limit reached")
	default:
		return nil
	}
}

// Execute loads the requested script, appends the script input data, sets up
// an execution environment, and interprets bytecodes until a termination
// condition is reached.
func (e *Engine) Execute(si state.ScriptInput) (totalCost int, err error) {
	// load wallet
	w, err := e.state.LoadWallet(si.WalletID)
	if err != nil {
		return
	}

	// initialize execution environment
	env := scriptEnv{
		script:   append(w.Script, si.Input...),
		dptr:     len(w.Script),
		wallet:   &w,
		engine:   e,
		deadline: si.Deadline,
		// these values will likely be stored as part of the wallet
		instBalance: maxInstructions,
		costBalance: 10000,
		memUsage:    0,
	}

	// run script
	if debug {
		fmt.Println("executing script:", env.script)
	}
	if err = env.run(); err != nil {
		fmt.Println("script execution failed:", err)
	}

	e.state.SaveWallet(w)
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
		if err := env.deductResources(op); err != nil {
			return err
		}

		// read bytes into argument array and advance env.iptr
		fnArgs := make([]byte, op.argBytes)
		env.iptr++
		env.iptr += copy(fnArgs, env.script[env.iptr:])

		// call associated opcode function and check for error
		if err := op.fn(env, fnArgs); err != nil {
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
			fmt.Println("\tstack:", env.stack.print())
			//b := make([]byte, 20)
			//copy(b, env.registers[1])
			//fmt.Println("    register 1:", len(env.registers[1]), b)
			//b = make([]byte, 20)
			//copy(b, env.registers[2])
			//fmt.Println("    register 2:", len(env.registers[2]), b)
		}
	}
}
