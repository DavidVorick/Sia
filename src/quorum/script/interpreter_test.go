package script

import (
	"bytes"
	"testing"
)

func TestOpCodes(t *testing.T) {
	// test jump
	_, err := ExecuteScript([]byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		0x0B,       // compare 2 == 3
		0x12, 0x0A, // if 2 == 3, jump ahead 10 instructions (causing an error)
		0xFF,
	})
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test store/load
	_, err = ExecuteScript([]byte{
		0x01, 0x02, // push 2
		0x03,       // dup 2
		0x14, 0x01, // store 2 in register 1
		0x15, 0x01, // load 2 from register 1
		0x0B,       // compare 2 == 2
		0x0F,       // boolean NOT
		0x12, 0x0A, // if 2 != 2, jump ahead 10 instructions (causing an error)
		0xFF,
	})
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test invalid script
	_, err = ExecuteScript([]byte{
		0x00,
		0x01, 0xAA,
		0x01, 0x1B,
		0x05,
		0x02,
		0x05,
		0xFF,
	})
	if err == nil {
		t.Fatal("expected stack empty error")
	}
}

func BenchmarkInterpreter(b *testing.B) {
	// push a thousand bytes and perform various operations on them
	script := bytes.Repeat([]byte{0x01, 0xAA}, 1000)
	script = append(script, bytes.Repeat([]byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, 100)...)
	script = append(script, 0xFF)
	for n := 0; n < b.N; n++ {
		ExecuteScript(script)
	}
}
