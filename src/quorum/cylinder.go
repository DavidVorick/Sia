package quorum

import (
	"bytes"
	"common"
	"encoding/gob"
	"fmt"
	"siacrypto"
)

// a CID, or Cylinder ID, is the global logical address of a batch on Sia. A CID
// has no relationship to where on disk or in the AA tree the batch is stored.
// To perform lookups from a CID to the disk location of a CID, a map musst be
// used.
type CID int // not exactly sure what CID will end up looking like

// A cylinder is the set of 128 corresponding batches in a quorum.
type Cylinder struct {
	Hash      siacrypto.TruncatedHash
	RingPairs int
	RingAtoms []int
	RingMList []int
	CID       CID
}

// A cylinderMap maps CIDs to their cylinder object within the cylinderTree
type cylinderMap map[CID]batch

type AllocateCylinder struct {
	// a wallet to control the cylinder

	Cylinder Cylinder
}

func (a AllocateCylinder) process(p *Participant) {
	println("processing cylinder")
	// check that CID is not already taken
	_, exists := p.quorum.cylinderMap[a.Cylinder.CID]
	if exists {
		// error
		return
	}

	// check that the cylinder is valid
	if a.Cylinder.RingPairs == 0 {
		if len(a.Cylinder.RingAtoms) != 1 {
			// error
			return
		}
	} else {
		if len(a.Cylinder.RingAtoms) != 2*a.Cylinder.RingPairs {
			// error
			return
		}
	}

	if len(a.Cylinder.RingAtoms) != len(a.Cylinder.RingMList) {
		// error
		return
	}

	// Calculate weight of new cylinder
	weight := 8                    // 8 atoms for nonredundant error detection
	weight += a.Cylinder.RingPairs // 1 atom of error detection for each RingPair
	for _, length := range a.Cylinder.RingAtoms {
		weight += length // the atoms in each ring
	}

	// verify that there's enough room on disk for the new cylinder
	if p.quorum.cylinderTreeHead.weight+weight > common.AtomsPerStack {
		// error
		return
	}

	// Place cylinder into cylinderMap for lookups
	p.quorum.cylinderMap[a.Cylinder.CID] = &a.Cylinder

	// Create a new cylinderNode and insert into the cylinderTree
	cn := new(cylinderNode)
	cn.weight = weight
	cn.data = &a.Cylinder
	p.quorum.insert(cn)
	println("processed")
}

func (a *AllocateCylinder) GobEncode() (gobAC []byte, err error) {
	if a == nil {
		return nil, nil
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(a.Cylinder)
	if err != nil {
		return
	}
	err = encoder.Encode(a.Cylinder.CID)
	if err != nil {
		return
	}
	gobAC = w.Bytes()
	return
}

func (a *AllocateCylinder) GobDecode(gobAC []byte) (err error) {
	if a == nil {
		err = fmt.Errorf("Cannot decode into nil object")
		return
	}

	r := bytes.NewBuffer(gobAC)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&a.Cylinder)
	if err != nil {
		return
	}
	err = decoder.Decode(&a.Cylinder.CID)
	if err != nil {
		return
	}

	return
}
