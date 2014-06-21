package quorum

import (
	"siacrypto"
)

type upload struct {
	sectorID              string
	requiredConfirmations byte
	receivedConfirmations [QuorumSize]bool
	hashSet               [QuorumSize]siacrypto.Hash
	hash                  siacrypto.Hash
	weight                uint16
	deadline              uint32
	counter               uint64
}

type UploadAdvancement struct {
	SectorID  string
	Index     byte
	Sibling   byte
	Signature siacrypto.Signature
}

func (u *upload) handleEvent(q *Quorum) {
	if q.uploads[u.sectorID] == nil {
		return
	}
	var i int
	for i = 0; i < len(q.uploads[u.sectorID]); i++ {
		if q.uploads[u.sectorID][i].hash == u.hash {
			break
		}
	}

	if i == len(q.uploads[u.sectorID]) {
		return
	}
	q.clearUploads(u.sectorID, i)
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

func (q *Quorum) advanceUpload(u *UploadAdvancement) {
	// mark the sibling sibling as having completed the upload
	// then check if the upload is ready to have complete() called
	if q.uploads[u.SectorID] == nil {
		return
	}
	if byte(len(q.uploads[u.SectorID])) < u.Index {
		return
	}
	if q.siblings[u.Sibling] == nil {
		return
	}
	if q.uploads[u.SectorID][u.Index].receivedConfirmations[u.Sibling] == true {
		return
	}

	// verify that the signature belongs to the sibling
	advanceBytes := make([]byte, 10)
	copy(advanceBytes, []byte(u.SectorID))
	advanceBytes[8] = u.Index
	advanceBytes[9] = u.Sibling
	verified := q.siblings[u.Sibling].publicKey.Verify(&siacrypto.SignedMessage{
		Signature: u.Signature,
		Message:   advanceBytes,
	})
	if !verified {
		return
	}

	q.uploads[u.SectorID][u.Index].receivedConfirmations[u.Sibling] = true
	q.uploads[u.SectorID][u.Index].requiredConfirmations -= 1
	if q.uploads[u.SectorID][u.Index].requiredConfirmations == 0 {
		// completeUpload()
	}
}

func (q *Quorum) completeUpload(sectorID string) {
	// the upload at the first index is popped from the chain (just do a reslice), and the hash is moved over
	// costs are refunded where possible
}
