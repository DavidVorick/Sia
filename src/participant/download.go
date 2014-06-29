package participant

func (p *Participant) DownloadSegment(id quorum.WalletID, segment *[]byte) (err error) {
	filename := p.quorum.SectorFilename
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return
	}

	*segment = make([]byte, info.Size())
	_, err := file.Read(*segment)
	return
}
