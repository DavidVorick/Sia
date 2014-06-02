package quorum

/*import (
	"testing"
)

// countReachableNodes iterates through a batchTree and counts how many nodes
// are reachable from the input, which is assumed to be the parent.
//
// countReachableNodes does not check for cycles and will get caught in an
// infinite loop.
func countReachableNodes(bn *wallet) (i int) {
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
	q := new(Quorum)
	q.walletTreeHead = new(wallet)
	q.walletTreeHead.weight = 1
	child0 := new(wallet)
	child0.weight = 5
	child1 := new(wallet)
	child1.weight = 9
	child2 := new(wallet)
	child2.weight = 6
	child3 := new(wallet)
	child3.weight = 24
	child4 := new(wallet)
	child4.weight = 55

	// insert children into batchTree
	q.insert(child0)
	q.insert(child1)
	q.insert(child2)
	q.insert(child3)
	q.insert(child4)

	reachableNodes := countReachableNodes(q.walletTreeHead)
	if reachableNodes != 6 {
		t.Error("After insertion, wrong number of nodes counted as reachable:", reachableNodes)
	}

	// verify that aggregate values are correct
	if q.walletTreeHead.weight != 100 {
		t.Error("cylinderTree weight not updating correctly with insert:", q.walletTreeHead.weight)
	}
	if q.walletTreeHead.children != 5 {
		t.Error("cylinderTree children not updating correctly with insert:", q.walletTreeHead.children)
	}

	// delete children from aggregate tree
	q.delete(child0)
	q.delete(child1)
	q.delete(child2)
	q.delete(child3)
	q.delete(child4)

	reachableNodes = countReachableNodes(q.walletTreeHead)
	if reachableNodes != 1 {
		t.Error("After deletion, wrong number of nodes counted as reachable:", reachableNodes)
	}

	// verify that aggregate values are correct
	if q.walletTreeHead.weight != 1 {
		t.Error("cylinderTree weight not updating correctly with delete:", q.walletTreeHead.weight)
	}
	if q.walletTreeHead.children != 0 {
		t.Error("cylinderTree children not updating correctly with delete:", q.walletTreeHead.children)
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

	// fill out sectors to give weights to each sector
	q.walletTreeHead.data = new(Cylinder)
	q.walletTreeHead.data.RingAtoms = make([]int, 1)
	q.walletTreeHead.data.RingAtoms[0] = 1
	child0.data = new(Cylinder)
	child0.data.RingAtoms = make([]int, 2)
	child0.data.RingAtoms[0] = 2
	child0.data.RingAtoms[1] = 3
	q.insert(child0)
	child1.data = new(Cylinder)
	child1.data.RingAtoms = make([]int, 2)
	child1.data.RingAtoms[0] = 4
	child1.data.RingAtoms[1] = 5
	q.insert(child1)
	child2.data = new(Cylinder)
	child2.data.RingAtoms = make([]int, 1)
	child2.data.RingAtoms[0] = 6
	q.insert(child2)
	child3.data = new(Cylinder)
	child3.data.RingAtoms = make([]int, 3)
	child3.data.RingAtoms[0] = 7
	child3.data.RingAtoms[1] = 8
	child3.data.RingAtoms[2] = 9
	q.insert(child3)
	child4.data = new(Cylinder)
	child4.data.RingAtoms = make([]int, 1)
	child4.data.RingAtoms[0] = 55
	q.insert(child4)

	reachableNodes = countReachableNodes(q.walletTreeHead)
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

		old := selectionMap[randomBatch.RingAtoms[randomSector]]
		old += 1
		selectionMap[randomBatch.RingAtoms[randomSector]] = old
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
}*/
