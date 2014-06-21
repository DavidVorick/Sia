package quorum

import (
	"siacrypto"
)

type upload struct {
	requiredConfirmations byte
	receivedConfirmations [QuorumSize]bool
	hashSet               [QuorumSize]siacrypto.Hash
	hash                  siacrypto.Hash
	weight                uint16
	deadline              uint32
	counter               uint64
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

func (q *Quorum) clearUploads(sectorID string, i int) {
	// delete all uploads starting with the ith index
}
