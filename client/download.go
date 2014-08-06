package client

/*
import (
	"bytes"
	"consensus"
	"fmt"
	"io"
	"network"
	"os"
	"siaencoding"
	"state"
)

// download every segment
// call repair
// look at the padding value
// truncate the padding

func (c *Client) Download(id state.WalletID, destination string) {
	c.RetrieveSiblings()

	// figure out how many atoms this wallet has
	var err error
	var w state.Wallet
	for i := range c.siblings {
		if c.siblings[i] == nil {
			continue
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.siblings[i].Address,
			Proc: "Participant.DownloadWallet",
			Args: id,
			Resp: &w,
		})
		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	if w.SectorSettings.Atoms < 2 {
		fmt.Println("Wallet is not storing a file.")
		return
	}

	// one at a time, download atoms from every sibling until all atoms have been downloaded
	segments := make([][][]byte, w.SectorSettings.Atoms)
	for i := uint16(1); i < w.SectorSettings.Atoms; i++ {
		segments[i] = make([][]byte, state.QuorumSize)
		ad := consensus.AtomDownload{
			ID:           id,
			AtomIndicies: []uint16{i},
		}
		for j := range c.siblings {
			if c.siblings[j] == nil {
				continue
			}
			err := c.router.SendMessage(network.Message{
				Dest: c.siblings[j].Address,
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
	sector := make([][]byte, w.SectorSettings.Atoms)
	for i := uint16(1); i < w.SectorSettings.Atoms; i++ {
		// read through the downloaded atoms, build and indices
		inputSegments := make([]io.Reader, w.SectorSettings.K)
		indices := make([]byte, w.SectorSettings.K)
		var j int
		for k := range segments[i] {
			if segments[i][k] != nil {
				inputSegments[j] = bytes.NewBuffer(segments[i][k])
				indices[j] = byte(k)
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
		_, err := state.RSRecover(inputSegments, indices, recoveredAtom, w.SectorSettings.K)
		if err != nil {
			fmt.Println("Error while recovering file")
			fmt.Println(err)
			return
		}
		sector[i] = recoveredAtom.Bytes()
	}

	encodedPadding := sector[1][:4]
	padding := siaencoding.DecUint32(encodedPadding)
	if padding > uint32(state.AtomSize) {
		fmt.Println("unexpected and illegal padding value - probably not a file")
		return
	}
	sector[1] = sector[1][4:]
	sector[w.SectorSettings.Atoms-1] = sector[w.SectorSettings.Atoms-1][:padding]

	file, err := os.Create(destination)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := 1; i < len(sector); i++ {
		file.Write(sector[i])
	}
}
*/
