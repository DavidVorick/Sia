package common

import (
	"sync"
)

type ZeroNetwork struct {
	messages     []*Message
	messagesLock sync.RWMutex
}

func (z *ZeroNetwork) Address() (a Address) {
	return
}

func (z *ZeroNetwork) RegisterHandler(handler interface{}) (i Identifier) {
	return
}

func (z *ZeroNetwork) SendMessage(m *Message) error {
	z.messagesLock.Lock()
	z.messages = append(z.messages, m)
	z.messagesLock.Unlock()
	return nil
}

func (z *ZeroNetwork) SendAsyncMessage(m *Message) (c chan error) {
	z.messagesLock.Lock()
	z.messages = append(z.messages, m)
	z.messagesLock.Unlock()
	c = make(chan error, 1)
	c <- nil
	return
}

func (z *ZeroNetwork) Close() {
	return
}

func (z *ZeroNetwork) RecentMessage(i int) *Message {
	z.messagesLock.RLock()
	defer z.messagesLock.RUnlock()
	if i < len(z.messages) {
		return z.messages[i]
	}
	return nil
}

func NewZeroNetwork() *ZeroNetwork {
	return new(ZeroNetwork)
}
