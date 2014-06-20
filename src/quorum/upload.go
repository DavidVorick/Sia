package quorum

type upload struct {
	atoms    [256]bool
	priority uint32
	deadline uint32
	counter  uint64
}

func (u *upload) handleEvent() {
}

func (u *upload) expiration() uint32 {
	return u.deadline
}

func (u *upload) setCounter(c uint64) {
	u.counter = c
}

func (u *upload) fetchCounter() uint64 {
	return u.counter
}
