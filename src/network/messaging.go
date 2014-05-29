package network

// An Identifier uniquely identifies a participant on a host.
type Identifier byte

// An Address couples an Identifier with its network address.
type Address struct {
	ID   Identifier
	Host string
	Port int
}

// A Message is for sending requests over the network.
// It consists of an Address and an RPC. It is the MessageRouter's job to
// route a message to its intended destination.
type Message struct {
	Dest Address
	Proc string
	Args interface{}
	Resp interface{}
}

// A MessageRouter both transmits outgoing messages and processes incoming messages.
// It dispenses Identifiers to objects that register themselves on the server.
type MessageRouter interface {
	Address() Address
	RegisterHandler(interface{}) Identifier
	SendMessage(*Message) error
	SendAsyncMessage(*Message) chan error
	Close()
}
