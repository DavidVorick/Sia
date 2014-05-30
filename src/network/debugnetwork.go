package network

import (
	"sync"
)

type DebugNetwork struct {
	messages     []*Message
	messagesLock sync.RWMutex
	numHandlers  int
}

func (z *DebugNetwork) Address() Address {
	return Address{0, "localhost", 9988}
}

func (z *DebugNetwork) RegisterHandler(handler interface{}) Identifier {
	z.numHandlers++
	return Identifier(z.numHandlers)
}

func (z *DebugNetwork) SendMessage(m *Message) error {
	z.messagesLock.Lock()
	z.messages = append(z.messages, m)
	z.messagesLock.Unlock()
	return nil
}

func (z *DebugNetwork) SendAsyncMessage(m *Message) (c chan error) {
	z.messagesLock.Lock()
	z.messages = append(z.messages, m)
	z.messagesLock.Unlock()
	c = make(chan error, 1)
	c <- nil
	return
}

func (z *DebugNetwork) Close() {
	return
}

func (z *DebugNetwork) RecentMessage(i int) *Message {
	z.messagesLock.RLock()
	defer z.messagesLock.RUnlock()
	if i < len(z.messages) {
		return z.messages[i]
	}
	return nil
}

func NewDebugNetwork() *DebugNetwork {
	return new(DebugNetwork)
}
