package script

import (
	"bytes"
	"testing"
)

func TestOpCodes(t *testing.T) {
	s := new(Script)
	// test jump
	s.Input = []byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		0x0B,       // compare 2 == 3
		0x12, 0x0A, // if 2 == 3, jump ahead 10 instructions (causing an error)
		0xFF,
	}
	_, err := s.Execute(nil)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test store/load
	s.Input = []byte{
		0x01, 0x02, // push 2
		0x03,       // dup 2
		0x14, 0x01, // store 2 in register 1
		0x15, 0x01, // load 2 from register 1
		0x0B,       // compare 2 == 2
		0x0F,       // boolean NOT
		0x12, 0x0A, // if 2 != 2, jump ahead 10 instructions (causing an error)
		0xFF,
	}
	_, err = s.Execute(nil)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test invalid script
	s.Input = []byte{
		0x00,
		0x01, 0xAA,
		0x01, 0x1B,
		0x05,
		0x02,
		0x05,
		0xFF,
	}
	_, err = s.Execute(nil)
	if err == nil {
		t.Fatal("expected stack empty error")
	}
}

func BenchmarkInterpreter(b *testing.B) {
	s := new(Script)
	// push a thousand bytes and perform various operations on them
	s.Input = bytes.Repeat([]byte{0x01, 0xAA}, 1000)
	s.Input = append(s.Input, bytes.Repeat([]byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, 100)...)
	s.Input = append(s.Input, 0xFF)
	for n := 0; n < b.N; n++ {
		s.Execute(nil)
	}
}
