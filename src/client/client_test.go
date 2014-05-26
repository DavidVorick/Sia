package main

import (
	"common"
	"common/crypto"
	"common/erasure"
	"network"
	"testing"
)

type Server struct {
	seg common.Segment
}

func (s *Server) UploadSegment(seg common.Segment, arb *struct{}) error {
	s.seg = seg
	return nil
}

func (s *Server) DownloadSegment(hash crypto.Hash, seg *common.Segment) error {
	*seg = s.seg
	return nil
}

// TestRPCUploadSector tests the NewRPCServer and uploadFile functions.
// NewRPCServer must properly initialize a RPC server.
// uploadSector must succesfully distribute a Sector among a quorum.
// The uploaded Sector must be successfully reconstructed.
func TestRPCuploadSector(t *testing.T) {
	SectorDB = make(map[crypto.Hash]*common.RingHeader)

	// create RPCServer
	var err error
	router, err = network.NewRPCServer(9985)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer router.Close()

	// create quorum
	var q [common.QuorumSize]common.Address
	var shs [common.QuorumSize]Server
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		defer qrpc.Close()
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		q[i].ID = qrpc.RegisterHandler(&shs[i])
	}

	// create sector
	secData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	sec, err := common.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	// add sector to database
	k := common.QuorumSize / 2
	SectorDB[sec.Hash] = &common.RingHeader{
		Hosts:  q,
		Params: sec.CalculateParams(k),
	}

	// upload sector to quorum
	err = uploadSector(sec)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// rebuild file from first k segments
	var newRing []common.Segment
	for i := 0; i < k; i++ {
		newRing = append(newRing, shs[i].seg)
	}

	sec, err = erasure.RebuildSector(newRing, SectorDB[sec.Hash].Params)
	if err != nil {
		t.Fatal("Failed to rebuild file:", err)
	}

	// check hash
	rebuiltHash, err := crypto.CalculateHash(sec.Data)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}

// TestRPCdownloadSector tests the NewRPCServer and downloadSector functions.
// NewRPCServer must properly initialize a RPC server.
// downloadSector must successfully retrieve a Sector from a quorum.
// The downloaded Sector must match the original Sector.
func TestRPCdownloadSector(t *testing.T) {
	SectorDB = make(map[crypto.Hash]*common.RingHeader)

	// create sector
	secData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	sec, err := common.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	k := common.QuorumSize / 2
	params := sec.CalculateParams(k)

	// encode sector
	ring, err := erasure.EncodeRing(sec, params)
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
	var q [common.QuorumSize]common.Address
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		sh := new(Server)
		sh.seg = ring[i]
		q[i].ID = qrpc.RegisterHandler(sh)
	}

	// add sector to database
	SectorDB[sec.Hash] = &common.RingHeader{
		Hosts:  q,
		Params: params,
	}

	// download file from quorum
	sec, err = downloadSector(sec.Hash)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	// check hash
	rebuiltHash, err := crypto.CalculateHash(sec.Data)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}
