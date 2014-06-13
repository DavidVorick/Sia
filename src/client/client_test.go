package main

import (
	"network"
	"quorum"
	"siacrypto"
	"testing"
)

type Server struct {
	seg quorum.Segment
}

func (s *Server) UploadSegment(seg quorum.Segment, _ *struct{}) error {
	s.seg = seg
	return nil
}

func (s *Server) DownloadSegment(hash siacrypto.Hash, seg *quorum.Segment) error {
	*seg = s.seg
	return nil
}

// TestRPCUploadSector tests the NewRPCServer and uploadFile functions.
// NewRPCServer must properly initialize a RPC server.
// uploadSector must succesfully distribute a Sector among a quorum.
// The uploaded Sector must be successfully reconstructed.
func TestRPCuploadSector(t *testing.T) {
	SectorDB = make(map[siacrypto.Hash]*quorum.RingHeader)

	// create RPCServer
	var err error
	router, err = network.NewRPCServer(9985)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer router.Close()

	// create quorum
	var q [quorum.QuorumSize]network.Address
	var shs [quorum.QuorumSize]Server
	for i := 0; i < quorum.QuorumSize; i++ {
		q[i] = network.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		defer qrpc.Close()
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		q[i].ID = qrpc.RegisterHandler(&shs[i])
	}

	// create sector
	secData := siacrypto.RandomByteSlice(70000)
	sec, err := quorum.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	// add sector to database
	k := quorum.QuorumSize / 2
	SectorDB[sec.Hash] = &quorum.RingHeader{
		Hosts:  q,
		Params: sec.CalculateParams(k),
	}

	// upload sector to quorum
	err = uploadSector(sec)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// rebuild file from first k segments
	var newRing []quorum.Segment
	for i := 0; i < k; i++ {
		newRing = append(newRing, shs[i].seg)
	}

	sec, err = quorum.RebuildSector(newRing, SectorDB[sec.Hash].Params)
	if err != nil {
		t.Fatal("Failed to rebuild file:", err)
	}

	// check hash
	rebuiltHash := siacrypto.CalculateHash(sec.Data)
	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}

// TestRPCdownloadSector tests the NewRPCServer and downloadSector functions.
// NewRPCServer must properly initialize a RPC server.
// downloadSector must successfully retrieve a Sector from a quorum.
// The downloaded Sector must match the original Sector.
func TestRPCdownloadSector(t *testing.T) {
	SectorDB = make(map[siacrypto.Hash]*quorum.RingHeader)

	// create sector
	secData := siacrypto.RandomByteSlice(70000)
	sec, err := quorum.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	k := quorum.QuorumSize / 2
	params := sec.CalculateParams(k)

	// encode sector
	ring, err := quorum.EncodeRing(sec, params)
	if err != nil {
		t.Fatal("Failed to encode sector data:", err)
	}

	// create RPCServer
	router, err = network.NewRPCServer(9985)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer router.Close()

	// create quorum
	var q [quorum.QuorumSize]network.Address
	for i := 0; i < quorum.QuorumSize; i++ {
		q[i] = network.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		sh := new(Server)
		sh.seg = ring[i]
		q[i].ID = qrpc.RegisterHandler(sh)
	}

	// add sector to database
	SectorDB[sec.Hash] = &quorum.RingHeader{
		Hosts:  q,
		Params: params,
	}

	// download file from quorum
	sec, err = downloadSector(sec.Hash)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	// check hash
	rebuiltHash := siacrypto.CalculateHash(sec.Data)
	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}
