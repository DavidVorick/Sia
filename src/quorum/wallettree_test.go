package quorum

import (
	"fmt"
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

func printTree(w *walletNode) {
	if w == nil {
		return
	}

	fmt.Println(w)
	printTree(w.children[0])
	printTree(w.children[1])
}

// findRedViolatoins iterates through a red-black tree starting with w as the
// root and verifies that in no situations does a red node have another red
// node as a child.
func findRedViolations(w *walletNode) (violations int) {
	if w == nil {
		return
	}

	if w.isRed() {
		if w.children[0] != nil {
			if w.children[0].isRed() {
				violations += 1
			}
		}

		if w.children[1] != nil {
			if w.children[1].isRed() {
				violations += 1
			}
		}
	}

	violations += findRedViolations(w.children[0])
	violations += findRedViolations(w.children[1])
	return violations
}

// findBlackViolations iterates through a red-black tree starting with w as the
// root and makes sure that the height of black nodes in the tree is the same
// at all places in the tree.
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

// findSortViolaitons iterates through a red-black tree starting with w as the
// root and makes sure that the entire tree is sorted with no out-of-order
// elements.
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
		//println("VIOLATION")
		//println(w.id)
		//println(selfWeight)
		violations += 1
	}

	leftViolations := findWeightViolations(w.children[0])
	rightViolations := findWeightViolations(w.children[1])
	violations += leftViolations
	violations += rightViolations
	return
}

// checkViolations goes through a tree and verifies it as a proper Sia
// red-black tree. This means checking for red violations, black violations,
// sort violations, and then makes sure that all of the weights in the tree
// make sense. It does not count the number of reachable nodes because it
// doesn't know how many nodes are supposed to be reachable. checkViolations
// takes the testing.T object as input so it can call t.Error directly.
func checkViolations(id int, w *walletNode, t *testing.T) {
	redViolations := findRedViolations(w)
	if redViolations != 0 {
		t.Error(id, "- Got red violations:", redViolations)
	}

	_, blackViolations := findBlackViolations(w)
	if blackViolations != 0 {
		t.Error(id, "- Got black violations:", blackViolations)
	}

	sortViolations := findSortViolations(w)
	if sortViolations != 0 {
		t.Error(id, "- Got sort violations:", sortViolations)
	}

	weightViolations := findWeightViolations(w)
	if weightViolations != 0 {
		t.Fatal(id, "- Got weight violations:", weightViolations)
	}
}

// Extensive testing to make sure that there are no errors with the red-black
// tree. This test is particularly fleshed out because red-black trees are
// tricky and I made many mistakes in the first implementation.
func TestWalletTree(t *testing.T) {
	// create a quorum and add a bunch of children. 32 is enough children to
	// create 5 layers, which means all variables in insert are being used. After
	// each insertion all checks are made to verify that there have been no
	// violations.
	q := new(Quorum)
	n := 32
	for i := 0; i < n; i++ {
		newWallet := new(walletNode)
		newWallet.id = WalletID(i)
		newWallet.weight = i
		q.insert(newWallet)

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != i+1 {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}
		checkViolations(0, q.walletRoot, t)
	}

	// make sure that all 32 of the inserted nodes can be reached.
	for i := 0; i < n; i++ {
		w := q.retrieve(WalletID(i))
		if w == nil {
			t.Error("Unreachable Node:", i)
		}
	}

	// remove all 32 of the inserted nodes. After each removal, the full set of
	// tests are used to verify that the tree remains valid.
	for i := 0; i < n; i++ {
		q.remove(WalletID(i))

		w := q.retrieve(WalletID(i))
		if w != nil {
			t.Error("Maganed to retreive a removed wallet:", i)
		}

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != n-1-i {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}
		checkViolations(1, q.walletRoot, t)
	}

	if testing.Short() {
		t.Skip()
	}

	// insert a bunch of random nodes into the tree, each time verifying that the
	// tree is still valid.
	weights := make(map[uint64]bool)
	n = 1000
	for i := 0; i < n; i++ {
		found := false
		var weight int
		var err error
		for !found {
			weight, err = siacrypto.RandomInt(5000)
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
		if reachableNodes != i+1 {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}
		checkViolations(2, q.walletRoot, t)
	}

	// randomly insert or delete elements
	for i := 0; i < n; i++ {
		insertOrDelete, err := siacrypto.RandomInt(2)
		if err != nil {
			t.Fatal(err)
		}

		if insertOrDelete == 0 {
			// insert
			found := false
			var weight int
			var err error
			for !found {
				weight, err = siacrypto.RandomInt(5000)
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

			checkViolations(4, q.walletRoot, t)
		} else {
			// delete
			// turn the weights map into a slice
			/*var weightSlice []uint64
			for key, value := range weights {
				if value {
					weightSlice = append(weightSlice, key)
				}
			}

			i, err = siacrypto.RandomInt(len(weightSlice))
			if err != nil {
				t.Fatal(err)
			}
			q.remove(WalletID(weightSlice[i]))
			checkViolations(5, q.walletRoot, t)
			weights[weightSlice[i]] = false*/
		}
	}

	// turn the weights map into a slice
	var weightSlice []uint64
	for key, value := range weights {
		if value {
			weightSlice = append(weightSlice, key)
		}
	}

	// verify that all weights within the slice are reachable
	for _, value := range weightSlice {
		w := q.retrieve(WalletID(value))
		if w == nil {
			t.Error("Retrieval Error:", value)
		}
	}

	// shuffle the slice
	for i := range weightSlice {
		newIndex, err := siacrypto.RandomInt(len(weightSlice) - i)
		if err != nil {
			t.Fatal(err)
		}
		newIndex += i

		tmp := weightSlice[newIndex]
		weightSlice[newIndex] = weightSlice[i]
		weightSlice[i] = tmp
	}

	// remove the elements one at a time
	for i, v := range weightSlice {
		q.remove(WalletID(v))

		w := q.retrieve(WalletID(v))
		if w != nil {
			t.Error("Maganed to retreive a removed wallet:", v)
		}

		reachableNodes := countReachableNodes(q.walletRoot)
		if reachableNodes != len(weightSlice)-1-i {
			t.Error("Wrong number of reachable nodes:", reachableNodes)
		}
		checkViolations(3, q.walletRoot, t)
	}
}
