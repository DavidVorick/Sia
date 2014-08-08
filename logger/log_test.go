package logger

import (
	"bytes"
	"testing"
)

type testwriter struct {
	input    [][]byte
	expected [][]byte
	// Current expected input (indexes expected)
	current int
	// If all is well, this will be true, otherwise, this will be false
	good bool
}

func newtw(expected [][]byte) (tw *testwriter) {
	tw = new(testwriter)
	tw.expected = expected
	tw.good = true
	return
}

func (tw *testwriter) Write(b []byte) (int, error) {
	if bytes.Contains(b, tw.expected[tw.current]) {
		tw.input = append(tw.input, b)
	} else {
		// I find this line immensely ironic
		Error("Incorrect input ", string(b), " given, expected ", string(tw.expected[tw.current]))
		tw.good = false
	}
	tw.current = (tw.current + 1) % len(tw.expected)
	// The priority logger doesn't check the outputs of write, which
	// should be changed.
	return 0, nil
}

func (tw testwriter) Good() bool {
	return tw.good
}

func (tw testwriter) Sample() []byte {
	if len(tw.input) > 0 {
		return tw.input[0]
	}
	return nil
}

func TestLoggerOrder(t *testing.T) {
	expected := [][]byte{
		[]byte{'a', 'l', 'p', 'h', 'a'},
		[]byte{'b', 'e', 't', 'a'},
		[]byte{'g', 'a', 'm', 'm', 'a'},
	}

	writer := newtw(expected)
	logger := NewPriorityLog(writer, PstdFlags, false)

	SetGlobalFlags(Perror | Pwarning | Pinfo | Pdebug)

	// Test the instant output
	for i := 0; i < 100; i++ {
		for _, v := range expected {
			logger.Error(string(v))
		}
	}

	// Rest the delayed output
	for i := 0; i < 100; i++ {
		for _, v := range expected {
			logger.Info(string(v))
		}
	}
	for i := 0; i < 100; i++ {
		for _, v := range expected {
			logger.Debug(string(v))
		}
	}

	logger.LogStored()

	if !writer.Good() {
		t.Fatal("Issue in ordering of log entries")
	}

	if len(writer.input) == 0 {
		t.Fatal("Writer did not record anything")
	}
}

type voidwriter struct{}

func (vw *voidwriter) Write(b []byte) (int, error) {
	return 0, nil
}

func BenchmarkLogger(b *testing.B) {
	logger := NewPriorityLog(new(voidwriter), PstdFlags, false)
	for i := 0; i < b.N; i++ {
		logger.Error("Garbage")
	}
}
