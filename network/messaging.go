package network

// An Identifier uniquely identifies a participant on a host.
type Identifier byte

// An Address couples an Identifier with its network address.
type Address struct {
	Host string
	Port uint16
	ID   Identifier
}

// A Message is for sending requests over the network.
// It consists of an Address and an RPC.
type Message struct {
	Dest Address
	Proc string
	Args interface{}
	Resp interface{}
}
