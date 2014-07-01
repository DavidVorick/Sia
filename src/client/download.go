package client

import (
	"bytes"
	"fmt"
	"io"
	"network"
	"os"
	"participant"
	"quorum"
	"siaencoding"
)

// download every segment
// call repair
// look at the padding value
// truncate the padding

func (c *Client) Download(id quorum.WalletID, destination string) {
	c.RetrieveSiblings()

	// figure out how many atoms this wallet has
	var err error
	var w quorum.Wallet
	for i := range c.siblings {
		if c.siblings[i] == nil {
			continue
		}
		err = c.router.SendMessage(&network.Message{
			Dest: c.siblings[i].Address(),
			Proc: "Participant.DownloadWallet",
			Args: id,
			Resp: &w,
		})
		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Println("Could not download file - connectivity errors!")
		return
	}

	// one at a time, download atoms from every sibling until all atoms have been downloaded
	segments := make([][][]byte, w.SectorAtoms())
	for i := uint16(1); i < w.SectorAtoms(); i++ {
		segments[i] = make([][]byte, quorum.QuorumSize)
		ad := participant.AtomDownload{
			ID:           id,
			AtomIndicies: []uint16{i},
		}
		for j := range c.siblings {
			if c.siblings[j] == nil {
				continue
			}
			err := c.router.SendMessage(&network.Message{
				Dest: c.siblings[j].Address(),
				Proc: "Participant.DownloadSegment",
				Args: ad,
				Resp: &segments[i][j],
			})

			if err != nil {
				segments[i][j] = nil
			}
		}
	}

	// go through the downloaded segments and repair the atoms into a single sector
	sector := make([][]byte, w.SectorAtoms())
	for i := uint16(1); i < w.SectorAtoms(); i++ {
		// read through the downloaded atoms, build and indicies
		inputSegments := make([]io.Reader, w.SectorM())
		indicies := make([]byte, w.SectorM())
		var j int
		for k := range segments[i] {
			if segments[i][k] != nil {
				inputSegments[j] = bytes.NewBuffer(segments[i][k])
				indicies[j] = byte(k)
				j++
			}
			if j == len(inputSegments) {
				break
			}
		}

		if j != len(inputSegments) {
			fmt.Println("Unable to download sufficient pieces")
			return
		}

		recoveredAtom := new(bytes.Buffer)
		_, err := quorum.RSRecover(inputSegments, indicies, recoveredAtom, w.SectorM())
		if err != nil {
			fmt.Println("Error while recovering file")
			fmt.Println(err)
			return
		}
		sector[i] = recoveredAtom.Bytes()
	}

	encodedPadding := sector[1][:4]
	padding := siaencoding.DecUint32(encodedPadding)
	sector[1] = sector[1][4:]
	sector[w.SectorAtoms()-1] = sector[w.SectorAtoms()-1][:padding]

	file, err := os.Create(destination)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := 1; i < len(sector); i++ {
		file.Write(sector[i])
	}
}
