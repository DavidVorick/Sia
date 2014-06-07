package script

import (
	"bytes"
	"testing"
)

func TestOpCodes(t *testing.T) {
	s := new(Script)
	// test jump
	s.Block = []byte{
		0x01, 0x02, // push 2
		0x01, 0x03, // push 3
		// if 2 == 3, jump ahead 10 instructions (causing an error)
		0x16, 0x1F, 0x00, 0x0A,
		0xFF,
	}
	_, err := s.Execute(nil, nil)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test store/load
	s.Block = []byte{
		0x01, 0x02, // push 2
		0x04,       // dup 2
		0x21, 0x01, // store 2 in register 1
		0x22, 0x01, // load 2 from register 1
		// if 2 != 2, jump ahead 10 instructions (causing an error)
		0x17, 0x1F, 0x00, 0x0A,
		0xFF,
	}
	_, err = s.Execute(nil, nil)
	if err != nil {
		t.Fatal("wrong execution path taken:", err)
	}

	// test invalid scripts
	s.Block = []byte{
		0x01, 0xAA,
		0x06, 0xFF,
	}
	_, err = s.Execute(nil, nil)
	if err == nil {
		t.Fatal("expected stack empty error")
	}
	s.Block = []byte{
		0x11,
	}
	_, err = s.Execute(nil, nil)
	if err == nil {
		t.Fatal("expected missing argument error")
	}
	s.Block = nil
	_, err = s.Execute(nil, nil)
	if err == nil {
		t.Fatal("expected missing terminator error")
	}
}

func BenchmarkInterpreter(b *testing.B) {
	s := new(Script)
	// push a thousand bytes and perform various operations on them
	s.Block = bytes.Repeat([]byte{0x01, 0xAA}, 1000)
	s.Block = append(s.Block, bytes.Repeat([]byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, 100)...)
	s.Block = append(s.Block, 0xFF)
	for n := 0; n < b.N; n++ {
		s.Execute(nil, nil)
	}
}
