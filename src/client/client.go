package main

import (
	"fmt"
	"network"
	"quorum"
	"siacrypto"
)

// global variables
// (with apologies to Haskell)
var (
	router   network.MessageRouter
	SectorDB map[siacrypto.Hash]*quorum.RingHeader
)

// uploadSector splits a Sector into a Ring and distributes it across a quorum.
// It hashes each of the Ring's segments and stores the hashes in the SectorDB.
func uploadSector(sec *quorum.Sector) (err error) {
	// look up Sector in SectorDB
	rh := SectorDB[sec.Hash]
	if rh == nil {
		return fmt.Errorf("Sector not found in database")
	}

	// create ring
	ring, err := quorum.EncodeRing(sec, rh.Params)
	if err != nil {
		return
	}

	// calculate and store segment hashes
	for i := range ring {
		rh.SegHashes[i], err = siacrypto.CalculateHash(ring[i].Data)
		if err != nil {
			return
		}
	}

	// for now we just send segment i to host i
	// this may need to be randomized for security
	for i := range rh.Hosts {
		err = router.SendMessage(&network.Message{
			Dest: rh.Hosts[i],
			Proc: "Server.UploadSegment",
			Args: ring[i],
			Resp: nil,
		})
		if err != nil {
			return
		}
	}

	return
}

// downloadSector retrieves a Ring from the quorum it is stored on.
// It reconstructs the original Sector from the Ring.
func downloadSector(hash siacrypto.Hash) (sec *quorum.Sector, err error) {
	// look up Sector in SectorDB
	rh := SectorDB[hash]
	if rh == nil {
		err = fmt.Errorf("Sector not found in database")
		return
	}

	// send requests to each member of the quorum
	var segs []quorum.Segment
	for i := range rh.Hosts {
		var seg quorum.Segment
		sendErr := router.SendMessage(&network.Message{
			Dest: rh.Hosts[i],
			Proc: "Server.DownloadSegment",
			Args: rh.SegHashes[i],
			Resp: &seg,
		})
		if sendErr == nil {
			segs = append(segs, seg)
		} else {
			fmt.Println(sendErr)
		}
	}

	// rebuild file
	sec, err = quorum.RebuildSector(segs, rh.Params)
	return
}

func readQuorumAddresses() (q [quorum.QuorumSize]network.Address) {
	var input int
	for i := range q {
		fmt.Print("Please enter port number ", i, ": ")
		fmt.Scanln(&input)
		q[i] = network.Address{2, "localhost", input}
	}
	return
}

func generateSector(q [quorum.QuorumSize]network.Address) (s *quorum.Sector, err error) {
	if q[0].Port == 0 {
		err = fmt.Errorf("you must connect to a quorum first")
		return
	}
	s, err = quorum.NewSector(siacrypto.RandomByteSlice(70000))
	if err != nil {
		return
	}
	SectorDB[s.Hash] = &quorum.RingHeader{
		Hosts:  q,
		Params: s.CalculateParams(quorum.QuorumSize / 2),
	}
	return
}

func main() {
	router, _ = network.NewRPCServer(9989)
	defer router.Close()
	SectorDB = make(map[siacrypto.Hash]*quorum.RingHeader)
	var (
		input string
		q     [quorum.QuorumSize]network.Address
		s     *quorum.Sector
		h     siacrypto.Hash
		err   error
	)
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")
		case "j":
			fmt.Println("joining quorum")
			q = readQuorumAddresses()
			fmt.Println("connected to quorum")
		case "g":
			fmt.Println("generating Sector")
			s, err = generateSector(q)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("failed to generate Sector")
				break
			}
			h = s.Hash
			fmt.Println("created Sector with hash", h[:10])
		case "u":
			fmt.Println("uploading file")
			err = uploadSector(s)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("upload failed")
				break
			}
			fmt.Println("upload successful")
		case "d":
			fmt.Println("downloading file")
			rs, err := downloadSector(h)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("download failed")
				break
			}
			rh, err := siacrypto.CalculateHash(rs.Data)
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Println("download successful")
			fmt.Println("hash:", rh[:10])
		case "q":
			return
		}
	}
}
