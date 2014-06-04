package quorum

import (
	"siacrypto"
	"testing"
)

// countReachableNodes iterates through a batchTree and counts how many nodes
// are reachable from the input, which is assumed to be the parent.
//
// countReachableNodes does not check for cycles and will get caught in an
// infinite loop.
func countReachableNodes(w *walletNode) (i int) {
	if w == nil {
		return
	}

	i += 1
	i += countReachableNodes(w.children[0])
	i += countReachableNodes(w.children[1])
	return i
}

func findRedViolations(w *walletNode) (i int) {
	if w == nil {
		return
	}

	if w.isRed() {
		if w.children[0] != nil {
			if w.children[0].isRed() {
				i += 1
			}
		}

		if w.children[1] != nil {
			if w.children[1].isRed() {
				i += 1
			}
		}
	}

	i += findRedViolations(w.children[0])
	i += findRedViolations(w.children[1])
	return i
}

func findBlackViolations(w *walletNode) (height, violations int) {
	if w == nil {
		height = 1
		violations = 0
		return
	}

	leftHeight, leftViolations := findBlackViolations(w.children[0])
	rightHeight, rightViolations := findBlackViolations(w.children[1])

	violations += leftViolations
	violations += rightViolations

	if leftHeight != rightHeight {
		violations += 1
	}

	height = leftHeight
	if !w.isRed() {
		height += 1
	}

	return
}

func findSortViolations(w *walletNode) (violations int) {
	if w == nil {
		return
	}

	if w.children[0] != nil && w.children[0].id >= w.id {
		violations += 1
	}

	if w.children[1] != nil && w.children[1].id <= w.id {
		violations += 1
	}

	violations += findSortViolations(w.children[0])
	violations += findSortViolations(w.children[1])
	return
}

// findWeightViolations assumes that the id and the weight of the node are the
// same. This is not true in practice, but it holds throughout the testing.
// It's a simplification that makes testing easier as you don't have to
// remember which id was associated with which weight.
func findWeightViolations(w *walletNode) (violations int) {
	if w == nil {
		return
	}

	selfWeight := w.weight
	if w.children[0] != nil {
		selfWeight -= w.children[0].weight
	}
	if w.children[1] != nil {
		selfWeight -= w.children[1].weight
	}
	if uint64(selfWeight) != uint64(w.id) {
		violations += 1
	}

	leftViolations := findWeightViolations(w.children[0])
	rightViolations := findWeightViolations(w.children[1])
	violations += leftViolations
	violations += rightViolations
	return
}

func TestWalletTree(t *testing.T) {
	// create a quorum and add a bunch of children
	q := new(Quorum)
	for i := 0; i < 32; i++ {
		newWallet := new(walletNode)
		newWallet.id = WalletID(i)
		newWallet.weight = i
		q.insert(newWallet)

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != i+1 {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}

		redViolations := findRedViolations(q.walletRoot)
		if redViolations != 0 {
			t.Error("Got red violations:", redViolations)
		}

		_, blackViolations := findBlackViolations(q.walletRoot)
		if blackViolations != 0 {
			t.Error("Got black violations:", blackViolations)
		}

		sortViolations := findSortViolations(q.walletRoot)
		if sortViolations != 0 {
			t.Error("Got sort violations:", sortViolations)
		}

		weightViolations := findWeightViolations(q.walletRoot)
		if weightViolations != 0 {
			t.Error("Got weight violations:", weightViolations)
		}
	}

	for i := 0; i < 32; i++ {
		q.remove(WalletID(i))

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != 31-i {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}

		redViolations := findRedViolations(q.walletRoot)
		if redViolations != 0 {
			t.Error("Got red violations:", redViolations)
		}

		_, blackViolations := findBlackViolations(q.walletRoot)
		if blackViolations != 0 {
			t.Error("Got black violations:", blackViolations)
		}

		sortViolations := findSortViolations(q.walletRoot)
		if sortViolations != 0 {
			t.Error("Got sort violations:", sortViolations)
		}

		weightViolations := findWeightViolations(q.walletRoot)
		if weightViolations != 0 {
			t.Error("Got weight violations:", weightViolations)
		}
	}

	if testing.Short() {
		t.Skip()
	}

	weights := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		found := false
		var weight int
		for !found {
			weight, err := siacrypto.RandomInt(5000)
			if err != nil {
				t.Fatal(err)
			}

			if weights[uint64(weight)] == false {
				weights[uint64(weight)] = true
				found = true
			}
		}

		node := new(walletNode)
		node.id = WalletID(weight)
		node.weight = weight
		q.insert(node)

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != i {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}

		redViolations := findRedViolations(q.walletRoot)
		if redViolations != 0 {
			t.Error("Got red violations:", redViolations)
		}

		_, blackViolations := findBlackViolations(q.walletRoot)
		if blackViolations != 0 {
			t.Error("Got black violations:", blackViolations)
		}

		sortViolations := findSortViolations(q.walletRoot)
		if sortViolations != 0 {
			t.Error("Got sort violations:", sortViolations)
		}

		weightViolations := findWeightViolations(q.walletRoot)
		if weightViolations != 0 {
			t.Error("Got weight violations:", weightViolations)
		}
	}

	for i := 0; i < 100; i++ {
		// add or delete a random value
	}

	// some way to delete values at random... ?
}
