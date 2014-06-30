package participant

import (
	"fmt"
	"os"
	"quorum"
)

const (
	MaxAtomsPerDownload = 16
)

type AtomDownload struct {
	ID           quorum.WalletID
	AtomIndicies []uint16
}

func (p *Participant) DownloadSegment(ad AtomDownload, segment *[]byte) (err error) {
	if len(ad.AtomIndicies) > MaxAtomsPerDownload {
		ad.AtomIndicies = ad.AtomIndicies[:MaxAtomsPerDownload]
	}

	filename := p.quorum.SectorFilename(ad.ID)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	info, err := file.Stat()

	*segment = make([]byte, len(ad.AtomIndicies)*quorum.AtomSize)
	for i := range ad.AtomIndicies {
		seekTo := int64(quorum.AtomSize) * int64(1+ad.AtomIndicies[i])
		if seekTo+int64(quorum.AtomSize) < info.Size() {
			err = fmt.Errorf("Invalid index request: sector is not composed of that many atoms!")
			return
		}

		_, err = file.Seek(int64(quorum.AtomSize)*int64(1+ad.AtomIndicies[i]), 0)
		if err != nil {
			panic(err)
		}

		_, err := file.Read((*segment)[quorum.AtomSize*i : quorum.AtomSize*(i+1)])
		if err != nil {
			panic(err)
		}
	}
	return
}
