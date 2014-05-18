package network

import (
	"common"
	"testing"
)

// a simple message handler
// stores a received message
type TestStoreHandler struct {
	message string
}

func (tsh *TestStoreHandler) StoreMessage(message string, arb *struct{}) error {
	tsh.message = message
	return nil
}

func (tsh *TestStoreHandler) DoNothing(message string, arb *struct{}) error {
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
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer rpcs.Close()

	// add a message handler to the server
	tsh := new(TestStoreHandler)
	id := rpcs.RegisterHandler(tsh)

	// send a message
	m := &common.Message{
		common.Address{id, "localhost", 9987},
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
	async := rpcs.SendAsyncMessage(m)
	<-async.Done
	if async.Error != nil {
		t.Fatal("Failed to send message:", async.Error)
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
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer rpcs.Close()

	// add a message handler to the server
	tsh := new(TestStoreHandler)
	id := rpcs.RegisterHandler(tsh)

	// send a message
	m := &common.Message{
		common.Address{id, "localhost", 9987},
		"TestStoreHandler.DoNothing",
		"hello, world!",
		nil,
	}
	err = rpcs.SendMessage(m)
	if err == nil {
		t.Fatal("Error: SendMessage did not timeout")
	}

	// send a message asynchronously
	tsh.message = ""
	async := rpcs.SendAsyncMessage(m)
	<-async.Done
	if async.Error == nil {
		t.Fatal("Error: SendAsyncMessage did not timeout")
	}
}
