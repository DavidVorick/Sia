package participant

import (
	"fmt"
	"os"
	"quorum"
	"siacrypto"
	"siafiles"
)

type Conversion struct {
	offset uint16
	delta  []byte
}

type UploadDiff struct {
	Id            quorum.WalletID
	Hash          siacrypto.Hash
	ConversionSet []Conversion
}

func (p *Participant) SignUploadAdvancement(ua *quorum.UploadAdvancement) {
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

func (p *Participant) ReceieveDiff(ud UploadDiff, _ struct{}) (err error) {
	// find the parent in the quorum
	if !p.quorum.ConfirmUpload(ud.Id, ud.Hash) {
		err = fmt.Errorf("Upload is not found in the quorum")
		return
	}

	// copy the file for the wallet over to a temporary state
	sectorFilename := p.quorum.SectorFilename(ud.Id)
	sectorSize, err := siafiles.Size(sectorFilename)
	if err != nil {
		panic(err)
	}
	for i := range ud.ConversionSet {
		if int(ud.ConversionSet[i].offset)+len(ud.ConversionSet[i].delta) > int(sectorSize) {
			err = fmt.Errorf("offset out of bounds error")
			return
		}
	}

	uploadFilename := sectorFilename + "." + string(ud.Hash[:])
	siafiles.Copy(sectorFilename, uploadFilename)

	// write the diffs into the file
	uploadFile, err := os.OpenFile(uploadFilename, os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer uploadFile.Close()
	for i := range ud.ConversionSet {
		_, err = uploadFile.Seek(int64(ud.ConversionSet[i].offset), 0)
		if err != nil {
			panic(err)
		}
		_, err = uploadFile.Write(ud.ConversionSet[i].delta)
		if err != nil {
			panic(err)
		}
	}

	// compare the hash of the file with the hash in the uploadRequest
	newHash, err := p.quorum.MerkleCollapse(uploadFilename)
	if err != nil {
		panic(err)
	}
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
		SectorID: ud.Id,
		Hash:     ud.Hash,
		Sibling:  p.self.Index(),
	}
	p.SignUploadAdvancement(&ua)
	p.uploadAdvancementsLock.Lock()
	p.uploadAdvancements = append(p.uploadAdvancements, ua)
	p.uploadAdvancementsLock.Unlock()
	return
}
