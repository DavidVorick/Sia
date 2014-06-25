package participant

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siafiles"
)

type Conversion struct {
	Offset uint16
	Delta  []byte
}

type UploadDiff struct {
	ID            quorum.WalletID
	Hash          siacrypto.Hash
	ConversionSet []Conversion
}

func (p *Participant) signUploadAdvancement(ua *quorum.UploadAdvancement) {
	gobUA, err := ua.GobEncode()
	if err != nil {
		panic(err)
	}
	signedMessage, err := p.secretKey.Sign(gobUA[:quorum.UnsignedUploadAdvancementSize])
	if err != nil {
		panic(err)
	}
	ua.Signature = signedMessage.Signature
}

func (p *Participant) ReceieveDiff(ud UploadDiff, _ *struct{}) (err error) {
	// find the upload in the quorum
	if !p.quorum.ConfirmUpload(ud.ID, ud.Hash) {
		err = fmt.Errorf("Upload is not found in the quorum")
		return
	}

	// Make sure that all offsets point to valid locations within the sector
	sectorFilename := p.quorum.SectorFilename(ud.ID)
	sectorSize, err := siafiles.Size(sectorFilename)
	if err != nil {
		panic(err)
	}
	for i := range ud.ConversionSet {
		if int(ud.ConversionSet[i].Offset)+len(ud.ConversionSet[i].Delta) > int(sectorSize) {
			err = fmt.Errorf("offset out of bounds error")
			return
		}
	}

	// copy the file for the wallet over to a temporary state
	uploadFilename := sectorFilename + "." + string(ud.Hash[:])
	siafiles.Copy(sectorFilename, uploadFilename)

	// write the diffs into the file
	uploadFile, err := os.OpenFile(uploadFilename, os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer uploadFile.Close()
	for i := range ud.ConversionSet {
		_, err = uploadFile.Seek(int64(ud.ConversionSet[i].Offset), 0)
		if err != nil {
			panic(err)
		}
		_, err = uploadFile.Write(ud.ConversionSet[i].Delta)
		if err != nil {
			panic(err)
		}
	}

	// important to remember that the first atom is used for the hashes
	_, err = uploadFile.Seek(int64(quorum.AtomSize), 0)
	if err != nil {
		panic(err)
	}

	// compare the hash of the file with the hash in the uploadRequest
	newHash := p.quorum.MerkleCollapse(uploadFile)
	if newHash != ud.Hash {
		err = os.Remove(uploadFilename)
		if err != nil {
			panic(err)
		}
		err = fmt.Errorf("diff did not result in correct hash")
		return
	}

	// submit an upload advancement to the quorum
	ua := quorum.UploadAdvancement{
		ID:      ud.ID,
		Hash:    ud.Hash,
		Sibling: p.self.Index(),
	}
	p.signUploadAdvancement(&ua)
	p.uploadAdvancementsLock.Lock()
	p.uploadAdvancements = append(p.uploadAdvancements, ua)
	p.uploadAdvancementsLock.Unlock()
	return
}
