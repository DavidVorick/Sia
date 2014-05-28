package quorum

import (
	"testing"
)

// countReachableNodes iterates through a batchTree and counts how many nodes
// are reachable from the input, which is assumed to be the parent.
//
// countReachableNodes does not check for cycles and will get caught in an
// infinite loop.
func countReachableNodes(bn *cylinderNode) (i int) {
	if bn == nil {
		return
	}
	i += 1
	i += countReachableNodes(bn.left)
	i += countReachableNodes(bn.right)
	return i
}

func TestBatchTree(t *testing.T) {
	// create a parent node and children
	parent := new(cylinderNode)
	parent.weight = 1
	child0 := new(cylinderNode)
	child0.weight = 5
	child1 := new(cylinderNode)
	child1.weight = 9
	child2 := new(cylinderNode)
	child2.weight = 6
	child3 := new(cylinderNode)
	child3.weight = 24
	child4 := new(cylinderNode)
	child4.weight = 55

	// insert children into batchTree
	parent.insert(child0)
	parent.insert(child1)
	parent.insert(child2)
	parent.insert(child3)
	parent.insert(child4)

	reachableNodes := countReachableNodes(parent)
	if reachableNodes != 6 {
		t.Error("After insertion, wrong number of nodes counted as reachable:", reachableNodes)
	}

	// verify that aggregate values are correct
	if parent.weight != 100 {
		t.Error("parent.weight not updating correctly with insert:", parent.weight)
	}
	if parent.children != 5 {
		t.Error("parent.children not updating correctly with insert:", parent.children)
	}

	// delete children from aggregate tree
	parent.delete(child0)
	parent.delete(child1)
	parent.delete(child2)
	parent.delete(child3)
	parent.delete(child4)

	reachableNodes = countReachableNodes(parent)
	if reachableNodes != 1 {
		t.Error("After deletion, wrong number of nodes counted as reachable:", reachableNodes)
	}

	// verify that aggregate values are correct
	if parent.weight != 1 {
		t.Error("parent.weight not updating correctly with delete:", parent.weight)
	}
	if parent.children != 0 {
		t.Error("parent.children not updating correctly with delete:", parent.children)
	}

	// After being deleted from the batchTree, the weight of a batch should be
	// equal to its original weight
	if child0.weight != 5 {
		t.Error("child0 weight altered after insertion and deletion:", child0.weight)
	}
	if child1.weight != 9 {
		t.Error("child1 weight altered after insertion and deletion:", child1.weight)
	}
	if child2.weight != 6 {
		t.Error("child2 weight altered after insertion and deletion:", child2.weight)
	}
	if child3.weight != 24 {
		t.Error("child3 weight altered after insertion and deletion:", child3.weight)
	}
	if child4.weight != 55 {
		t.Error("child4 weight altered after insertion and deletion:", child4.weight)
	}

	// create a quorum to fetch random numbers from
	q := new(quorum)
	q.parent = parent

	// fill out sectors to give weights to each sector
	parent.data = new(batch)
	parent.data.sectorLengths = make([]int, 1)
	parent.data.sectorLengths[0] = 1
	child0.data = new(batch)
	child0.data.sectorLengths = make([]int, 2)
	child0.data.sectorLengths[0] = 2
	child0.data.sectorLengths[1] = 3
	parent.insert(child0)
	child1.data = new(batch)
	child1.data.sectorLengths = make([]int, 2)
	child1.data.sectorLengths[0] = 4
	child1.data.sectorLengths[1] = 5
	parent.insert(child1)
	child2.data = new(batch)
	child2.data.sectorLengths = make([]int, 1)
	child2.data.sectorLengths[0] = 6
	parent.insert(child2)
	child3.data = new(batch)
	child3.data.sectorLengths = make([]int, 3)
	child3.data.sectorLengths[0] = 7
	child3.data.sectorLengths[1] = 8
	child3.data.sectorLengths[2] = 9
	parent.insert(child3)
	child4.data = new(batch)
	child4.data.sectorLengths = make([]int, 1)
	child4.data.sectorLengths[0] = 55
	parent.insert(child4)

	reachableNodes = countReachableNodes(parent)
	if reachableNodes != 6 {
		t.Error("After second insertion, wrong number of nodes counted as reachable:", reachableNodes)
	}

	// when short testing, skip statistical tests and fuzzing tests
	if testing.Short() {
		t.Skip()
	}

	// get several random batches and sectors, and verify they are all valid
	selectionMap := make(map[int]int)
	for i := 0; i < 1000000; i++ {
		randomBatch, randomSector := q.randomSector()

		old := selectionMap[randomBatch.sectorLengths[randomSector]]
		old += 1
		selectionMap[randomBatch.sectorLengths[randomSector]] = old
	}

	if len(selectionMap) != 10 {
		t.Error("not all values being selected by randomSector()")
	}

	for key, value := range selectionMap {
		if key*10500 < value || key*9500 > value {
			t.Error("Selection map statistically disrupted for key, value:", key, value)
		}
	}

	// fuzzing, add a bunch of elements of random weights, each time checking the
	// whole tree for integrity. then randomly add or delete random elements each
	// time checking the whole tree for integrity. Finally, remove all remaining
	// elements each time checking the whole tree for integrity.
	//
	// Also select random batches every time and verify that each selected batch
	// is valid. Then check the distribution of batches against statistical models
	// and confirm that it appears to be mathematically random
	//
	// a separate data structure will be needed to know how much weight each
	// element is supposed to have
}
