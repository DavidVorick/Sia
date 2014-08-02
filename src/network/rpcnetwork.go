package network

import (
	"errors"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var timeout = time.Second * 5

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
// It is the caller's responsibility to close the TCP listener, via RPCServer.Close().
func NewRPCServer(port int) (rpcs *RPCServer, err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	// determine our public hostname
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		tcpServ.Close()
		return
	}
	host := strings.Split(conn.LocalAddr().String(), ":")[0]
	conn.Close()

	rpcs = &RPCServer{
		addr:     Address{host, port, 0},
		rpcServ:  rpc.NewServer(),
		listener: tcpServ,
		curID:    1, // ID 0 is reserved for the RPCServer itself
	}

	go rpcs.serverHandler()
	return
}

// Close closes the connection associated with the TCP server.
// This causes tcpServ.Accept() to return an err, ending the serverHandler process.
func (rpcs *RPCServer) Close() {
	rpcs.listener.Close()
}

// serverHandler runs in the background, accepting incoming RPCs, serving them, and closing the connection.
// It is automatically terminated when Close() is called.
func (rpcs *RPCServer) serverHandler() {
	for {
		conn, err := rpcs.listener.Accept()
		if err != nil {
			return
		}
		go rpcs.rpcServ.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

// Ping calls the Participant.Ping method on the specified address.
func (rpcs *RPCServer) Ping(a Address) error {
	conn, err := jsonrpc.Dial("tcp", net.JoinHostPort(a.Host, strconv.Itoa(a.Port)))
	if err != nil {
		return err
	}
	defer conn.Close()

	select {
	case call := <-conn.Go("Participant"+string(a.ID)+".Ping", struct{}{}, nil, nil).Done:
		return call.Error
	case <-time.After(timeout):
		return nil
	}
}

// SendMessage synchronously delivers a Message to its recipient and returns any errors.
// It times out after waiting for 'timeout' seconds.
func (rpcs *RPCServer) SendMessage(m Message) error {
	conn, err := jsonrpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
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
	case <-time.After(timeout):
		return errors.New("request timed out")
	}
}

// SendAsyncMessage (asynchronously) delivers a Message to its recipient.
// It returns a channel that will contain an error value when the request completes.
// Like SendMessage, it times out after 'timeout' seconds.
func (rpcs *RPCServer) SendAsyncMessage(m Message) chan error {
	errChan := make(chan error, 2)
	conn, err := jsonrpc.Dial("tcp", net.JoinHostPort(m.Dest.Host, strconv.Itoa(m.Dest.Port)))
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
		case <-time.After(timeout):
			errChan <- errors.New("request timed out")
			return
		}
	}()

	return errChan
}
