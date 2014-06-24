package quorum

import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
)

const (
	UploadAdvancementSize         = WalletIDSize + siacrypto.HashSize + 1 + siacrypto.SignatureSize
	UnsignedUploadAdvancementSize = UploadAdvancementSize - siacrypto.SignatureSize
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
	SectorID  WalletID
	Hash      siacrypto.Hash
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

func (u *UploadAdvancement) GobEncode() (gobUA []byte, err error) {
	if u == nil {
		err = fmt.Errorf("Cannot encode nil UploadAdvancement")
		return
	}

	gobUA = make([]byte, UploadAdvancementSize)
	copy(gobUA, u.SectorID.Bytes())
	offset := WalletIDSize
	copy(gobUA[offset:], u.Hash[:])
	offset += siacrypto.HashSize
	gobUA[offset] = u.Sibling
	offset += 1
	copy(gobUA[offset:], u.Signature[:])
	return
}

func (u *UploadAdvancement) GobDecode(gobUA []byte) (err error) {
	if u == nil {
		err = fmt.Errorf("Cannode decode into nil UploadAdvancement")
		return
	}

	u.SectorID = WalletID(siaencoding.DecUint64(gobUA[0:WalletIDSize])) // bad, use GobDecode
	offset := WalletIDSize
	copy(u.Hash[:], gobUA[offset:])
	offset += siacrypto.HashSize
	u.Sibling = gobUA[offset]
	offset += 1
	copy(u.Signature[:], gobUA[offset:])
	return
}

func (q *Quorum) ConfirmUpload(id WalletID, h siacrypto.Hash) bool {
	idString := string(id.Bytes())
	for i := range q.uploads[idString] {
		if q.uploads[idString][i].hash == h {
			return true
		}
	}

	return false
}

func (q *Quorum) clearUploads(sectorID string, i int) {
	// delete all uploads starting with the ith index
}

func (q *Quorum) advanceUpload(ua *UploadAdvancement) {
	// check that all the associated structures exist
	sectorIDString := string(ua.SectorID.Bytes())
	if q.uploads[sectorIDString] == nil {
		return
	}
	if q.siblings[ua.Sibling] == nil {
		// this should never happen
		return
	}

	// find the index associated with this hash
	var i int
	for i = range q.uploads[sectorIDString] {
		if q.uploads[sectorIDString][i].hash == ua.Hash {
			break
		}
	}

	// see if upload exists in quorum
	if i == len(q.uploads[sectorIDString]) {
		return
	}

	// see if this sibling has already confirmed this upload advancement
	if q.uploads[sectorIDString][i].receivedConfirmations[ua.Sibling] == true {
		return
	}

	// verify that the signature belongs to the sibling
	uaBytes, err := ua.GobEncode()
	if err != nil {
		panic(err)
	}
	verified := q.siblings[ua.Sibling].publicKey.Verify(&siacrypto.SignedMessage{
		Signature: ua.Signature,
		Message:   uaBytes[:UnsignedUploadAdvancementSize],
	})
	if !verified {
		return
	}

	q.uploads[sectorIDString][i].receivedConfirmations[ua.Sibling] = true
	q.uploads[sectorIDString][i].requiredConfirmations -= 1
	if q.uploads[sectorIDString][i].requiredConfirmations <= 0 && i == 0 {
		// copy the upload file over to the actual file
		sectorFilename := q.SectorFilename(ua.SectorID)
		uploadFilename := sectorFilename + "." + string(ua.Hash)
		err := os.Rename(uploadFilename, sectorFilename)
		if err != nil {
			panic(err)
		}

		// subtract the temporary atoms from the wallet
		err := q.updateWeight(ua.ID, -ua.weight)
		if err != nil {
			panic(err)
		}

		// take the upload out of the uploads array
		q.uploads[ua.SectorID] = q.uploads[sectorID][1:]
	}
}
