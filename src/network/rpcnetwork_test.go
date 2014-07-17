package network

import (
	"bytes"
	"testing"
	"time"
)

// a simple message handler
// stores a received message
type TestStoreHandler struct {
	message string
}

func (tsh *TestStoreHandler) StoreMessage(message string, _ *struct{}) error {
	tsh.message = message
	return nil
}

func (tsh *TestStoreHandler) BlockForever(message string, _ *struct{}) error {
	select {}
	return nil
}

// TestRPCSendMessage tests the NewRPCServer, RegisterHandler, and Send(Async)Message functions.
// NewRPCServer must properly initialize a RPC server.
// RegisterHandler must make an RPC available to the client.
// SendMessage and SendAsyncMessage must complete successfully.
func TestRPCSendMessage(t *testing.T) {
	// create RPCServer
	rpcs, err := NewRPCServer(9987)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer rpcs.Close()

	// add a message handler to the server
	tsh := new(TestStoreHandler)
	id := rpcs.RegisterHandler(tsh)

	// send a message
	m := &Message{
		Address{id, "localhost", 9987},
		"TestStoreHandler.StoreMessage",
		"hello, world!",
		nil,
	}
	err = rpcs.SendMessage(m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}

	if tsh.message != "hello, world!" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"" + tsh.message + "\"")
	}

	// send a message asynchronously
	tsh.message = ""
	errChan := rpcs.SendAsyncMessage(m)
	err = <-errChan
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}

	if tsh.message != "hello, world!" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"" + tsh.message + "\"")
	}
}

// TestRPCTimeout tests the timeout functionality of Send(Async)Message.
// During the test, a message is sent to a handler that does nothing with it.
// The sender should eventually timeout and return an error instead of continuing to wait.
func TestRPCTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// create RPCServer
	rpcs, err := NewRPCServer(9987)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer rpcs.Close()

	// add a message handler to the server
	tsh := new(TestStoreHandler)
	id := rpcs.RegisterHandler(tsh)

	// send a message
	m := &Message{
		Dest: Address{id, "localhost", 9987},
		Proc: "TestStoreHandler.BlockForever",
		Args: "hello, world!",
		Resp: nil,
	}
	err = rpcs.SendMessage(m)
	if err == nil {
		t.Fatal("Error: SendMessage did not timeout")
	}

	// send a message asynchronously
	tsh.message = ""
	errChan := rpcs.SendAsyncMessage(m)
	if <-errChan == nil {
		t.Fatal("Error: SendAsyncMessage did not timeout")
	}
}

// TestRPCScheduling tests the RPC server's ability to process multiple concurrent messages.
// It is crucial that heartbeat RPCs are not blocked by other calls, such as uploads/downloads.
// This test starts one large data transfer and then attempts to send multiple smaller RPC messages.
// The smaller messages should arrive in a timely fashion despite the ongoing data transfer.
func TestRPCScheduling(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// create RPCServers
	rpcs1, err := NewRPCServer(9987)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer rpcs1.Close()
	rpcs2, err := NewRPCServer(9986)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer rpcs2.Close()

	// add a mesage handler to the servers
	tsh1 := new(TestStoreHandler)
	id1 := rpcs1.RegisterHandler(tsh1)
	tsh2 := new(TestStoreHandler)
	id2 := rpcs2.RegisterHandler(tsh2)

	// begin transferring large payload
	largeChan := rpcs2.SendAsyncMessage(&Message{
		Dest: Address{id1, "localhost", 9987},
		Proc: "TestStoreHandler.StoreMessage",
		Args: string(bytes.Repeat([]byte{0x10}, 1<<20)),
		Resp: nil,
	})

	// begin transferring small payload
	smallChan := rpcs1.SendAsyncMessage(&Message{
		Dest: Address{id2, "localhost", 9986},
		Proc: "TestStoreHandler.StoreMessage",
		Args: string(bytes.Repeat([]byte{0x10}, 1<<16)),
		Resp: nil,
	})

	// poll until both transfers complete
	var t1, t2 time.Time
	for i := 0; i < 2; i++ {
		select {
		case <-largeChan:
			t1 = time.Now()
		case <-smallChan:
			t2 = time.Now()
		}
	}

	if t2.After(t1) {
		t.Fatal("small transfer was blocked by large transfer")
	}
}
