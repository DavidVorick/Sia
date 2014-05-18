package network

import (
	"common"
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
	addr     common.Address
	rpcServ  *rpc.Server
	listener net.Listener
	curID    common.Identifier
}

func (rpcs *RPCServer) Address() common.Address {
	return rpcs.addr
}

// RegisterHandler registers a message handler to the RPC server.
// The handler is assigned an Identifier, which is returned to the caller.
// The Identifier is appended to the service name before registration.
func (rpcs *RPCServer) RegisterHandler(handler interface{}) (id common.Identifier) {
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
		addr:     common.Address{0, "localhost", port},
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
func (rpcs *RPCServer) SendMessage(m *common.Message) error {
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
	if err != nil {
		return err
	}

	// add identifier to service name
	name := strings.Replace(m.Proc, ".", string(m.Dest.ID)+".", 1)

	// send message
	select {
	case call := <-conn.Go(name, m.Args, m.Resp, nil).Done:
		return call.Error
	case <-time.After(common.StepDuration / 2):
		return errors.New("request timed out")
	}
}

// SendAsyncRPCMessage (asynchronously) delivers a Message to its recipient.
// It returns a *Call, which contains the fields "Done channel" and "Error error".
func (rpcs *RPCServer) SendAsyncMessage(m *common.Message) *rpc.Call {
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
	d := make(chan *rpc.Call, 1)
	if err != nil {
		// make a dummy *Call
		errCall := &rpc.Call{"", nil, nil, err, d}
		errCall.Done <- nil
		return errCall
	}
	// add identifier to service name
	name := strings.Replace(m.Proc, ".", string(m.Dest.ID)+".", 1)
	return conn.Go(name, m.Args, m.Resp, d)
}
