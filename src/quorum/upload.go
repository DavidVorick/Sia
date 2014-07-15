package quorum

/* import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
	"siafiles"
)

const (
	UploadAdvancementSize         = WalletIDSize + siacrypto.HashSize + 1 + siacrypto.SignatureSize
	UnsignedUploadAdvancementSize = UploadAdvancementSize - siacrypto.SignatureSize
)

type upload struct {
	id                    WalletID
	requiredConfirmations byte
	receivedConfirmations [QuorumSize]bool
	hashSet               [QuorumSize]siacrypto.Hash
	hash                  siacrypto.Hash
	weight                uint16
	deadline              uint32
	counter               uint64
}

type UploadAdvancement struct {
	ID        WalletID
	Hash      siacrypto.Hash
	Sibling   byte
	Signature siacrypto.Signature
}

func (u *upload) handleEvent(q *Quorum) {
	if q.uploads[u.id] == nil {
		return
	}
	var i int
	for i = 0; i < len(q.uploads[u.id]); i++ {
		if q.uploads[u.id][i].hash == u.hash {
			break
		}
	}

	if i == len(q.uploads[u.id]) {
		return
	}
	q.clearUploads(u.id, i)
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
	copy(gobUA, u.ID.Bytes())
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

	u.ID = WalletID(siaencoding.DecUint64(gobUA[0:WalletIDSize])) // bad, use GobDecode
	offset := WalletIDSize
	copy(u.Hash[:], gobUA[offset:])
	offset += siacrypto.HashSize
	u.Sibling = gobUA[offset]
	offset += 1
	copy(u.Signature[:], gobUA[offset:])
	return
}

func (q *Quorum) ConfirmUpload(id WalletID, h siacrypto.Hash) bool {
	for i := range q.uploads[id] {
		if q.uploads[id][i].hash == h {
			return true
		}
	}

	return false
}

func (q *Quorum) UploadExpectedHash(id WalletID, h siacrypto.Hash, index byte) (expected siacrypto.Hash) {
	for i := range q.uploads[id] {
		if q.uploads[id][i].hash == h {
			expected = q.uploads[id][i].hashSet[index]
			return
		}
	}
	return
}

func (q *Quorum) clearUploads(id WalletID, i int) {
	// if there are no uploads, mission accomplished
	if q.uploads[id] == nil {
		return
	}
	if i == 0 {
		q.uploads[id] = nil
		return
	}

	// delete all uploads starting with the ith index
	for i = i; i < len(q.uploads[id]); i++ {
		// delete the file associated with the upload
		sectorFilename := q.SectorFilename(id)
		uploadFilename := sectorFilename + "." + siafiles.SafeFilename(q.uploads[id][i].hash[:])
		err := os.Remove(uploadFilename)
		if err != nil {
			panic(err)
		}

		// subtract temporary atoms from the wallet
		err = q.updateWeight(id, int(-q.uploads[id][i].weight))
		if err != nil {
			panic(err)
		}
	}

	// remove the uploads from the slice
	q.uploads[id] = q.uploads[id][:i]
}

func (q *Quorum) AdvanceUpload(ua *UploadAdvancement) {
	// check that all the associated structures exist
	if q.uploads[ua.ID] == nil {
		return
	}
	if q.siblings[ua.Sibling] == nil {
		// this should never happen
		return
	}

	// find the index associated with this hash
	var i int
	for i = range q.uploads[ua.ID] {
		if q.uploads[ua.ID][i].hash == ua.Hash {
			break
		}
	}

	// see if upload exists in quorum
	if i == len(q.uploads[ua.ID]) {
		return
	}

	// see if this sibling has already confirmed this upload advancement
	if q.uploads[ua.ID][i].receivedConfirmations[ua.Sibling] == true {
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

	q.uploads[ua.ID][i].receivedConfirmations[ua.Sibling] = true
	q.uploads[ua.ID][i].requiredConfirmations -= 1
	if q.uploads[ua.ID][i].requiredConfirmations <= 0 && i == 0 {
		// copy the upload file over to the actual file
		sectorFilename := q.SectorFilename(ua.ID)
		uploadFilename := sectorFilename + "." + siafiles.SafeFilename(ua.Hash[:])
		err := os.Rename(uploadFilename, sectorFilename)
		if err != nil {
			panic(err)
		}

		// subtract the temporary atoms from the wallet
		err = q.updateWeight(ua.ID, 0-int(q.uploads[ua.ID][0].weight))
		if err != nil {
			panic(err)
		}

		// take the upload out of the uploads array
		q.uploads[ua.ID] = q.uploads[ua.ID][1:]
	}
} */
