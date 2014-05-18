package quorum

import (
	"os"
)

// a BID, or Batch ID, is the global logical address of a batch on Sia. A BID
// has no relationship to where on disk or in the AA tree the batch is stored.
// To perform lookups from a BID to the disk location of a BID, a batchMap must
// be used.
type BID [32]byte // not exactly sure what BID will end up looking like

// A batch is a group of sectors that is error-corrected together, and is the
// smallest unit that will move around on the network.
type batch struct {
	bid           *BID
	file          *os.File
	node          *batchNode
	sectorLengths []int
}

// A batchMap maps BIDs to their batch object
type batchMap map[BID]batch

// batchNode is the basic element in the AA tree that enables efficient
// selection of random sectors for verification
type batchNode struct {
	parent *batchNode
	left   *batchNode
	right  *batchNode

	children int
	weight   int
}

// addBatch takes a batch object that is not yet in the batchTree and puts it
// into the batchTree
func (parent *batchNode) insert(batch *batch) {
	batch.node = new(batchNode)
	batchWeight := 0
	for _, value := range batch.sectorLengths {
		batchWeight += value
	}

	// insert the node into the lightest-weight half of the parent
	currentNode := parent
	for currentNode.children > 1 {
		if currentNode.left.children < currentNode.right.children {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}
	}
	if currentNode.left == nil {
		currentNode.left = batch.node
	} else if currentNode.left == nil {
		currentNode.right = batch.node
	}
	batch.node.parent = currentNode

	// update the aggregate values of all parents
	for currentNode != nil {
		currentNode.children += 1
		currentNode.weight += batchWeight
		currentNode = currentNode.parent
	}
}
