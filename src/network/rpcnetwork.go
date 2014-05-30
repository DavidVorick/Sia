package network

import (
	"errors"
	"net"
	"net/rpc"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RPCServer is a MessageRouter that communicates using RPC over TCP.
type RPCServer struct {
	addr     Address
	rpcServ  *rpc.Server
	listener net.Listener
	curID    Identifier
}

func (rpcs *RPCServer) Address() Address {
	return rpcs.addr
}

// RegisterHandler registers a message handler to the RPC server.
// The handler is assigned an Identifier, which is returned to the caller.
// The Identifier is appended to the service name before registration.
func (rpcs *RPCServer) RegisterHandler(handler interface{}) (id Identifier) {
	id = rpcs.curID
	name := reflect.Indirect(reflect.ValueOf(handler)).Type().Name() + string(id)
	rpcs.rpcServ.RegisterName(name, handler)
	rpcs.curID++
	return
}

// NewRPCServer creates and initializes a server that listens for TCP connections on a specified port.
// It then spawns a serverHandler with a specified message.
// It is the callers's responsibility to close the TCP connection, via RPCServer.Close().
func NewRPCServer(port int) (rpcs *RPCServer, err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	rpcs = &RPCServer{
		addr:     Address{0, "localhost", port},
		rpcServ:  rpc.NewServer(),
		listener: tcpServ,
		curID:    1, // ID 0 is reserved for the RPCServer itself
	}

	go rpcs.serverHandler()
	return
}

// Close closes the connection associated with the TCP server.
// This causes tcpServ.Accept() to return an err, ending the serverHandler process
func (rpcs *RPCServer) Close() {
	rpcs.listener.Close()
}

// serverHandler accepts incoming RPCs, serves them, and closes the connection.
func (rpcs *RPCServer) serverHandler() {
	for {
		conn, err := rpcs.listener.Accept()
		if err != nil {
			return
		} else {
			go func() {
				rpcs.rpcServ.ServeConn(conn)
				conn.Close()
			}()
		}
	}
}

// SendRPCMessage (synchronously) delivers a Message to its recipient and returns any errors.
// It times out after waiting for half the step duration.
func (rpcs *RPCServer) SendMessage(m *Message) error {
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
	if err != nil {
		return err
	}
	defer conn.Close()

	// add identifier to service name
	name := strings.Replace(m.Proc, ".", string(m.Dest.ID)+".", 1)

	// send message
	select {
	case call := <-conn.Go(name, m.Args, m.Resp, nil).Done:
		return call.Error
	case <-time.After(time.Second * 5):
		return errors.New("request timed out")
	}
}

// SendAsyncRPCMessage (asynchronously) delivers a Message to its recipient.
// It returns a channel that will contain an error value when the request completes.
func (rpcs *RPCServer) SendAsyncMessage(m *Message) chan error {
	errChan := make(chan error, 2)
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
	if err != nil {
		errChan <- err
		return errChan
	}

	// add identifier to service name
	name := strings.Replace(m.Proc, ".", string(m.Dest.ID)+".", 1)

	// send message
	go func() {
		defer conn.Close()
		select {
		case call := <-conn.Go(name, m.Args, m.Resp, nil).Done:
			errChan <- call.Error
			return
		case <-time.After(time.Second * 5):
			errChan <- errors.New("request timed out")
			return
		}
	}()

	return errChan
}
