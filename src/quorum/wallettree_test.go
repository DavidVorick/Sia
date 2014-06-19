package quorum

import (
	"fmt"
	"siacrypto"
	"testing"
)

// countReachableNodes iterates through the rbw tree and counts how many nodes
// can be reached from the input node. It does not check for cycles and will
// get caught in an infinite loop if there are any cycles.
func countReachableNodes(w *walletNode) (i int) {
	if w == nil {
		return
	}

	i += 1
	i += countReachableNodes(w.children[0])
	i += countReachableNodes(w.children[1])
	return i
}

// printTree iterates through the rbw tree and prints each element. Because the
// tree is sorted and has weights, and is printed in DFS order, it's pretty
// trivial to reconstruct the tree without directly being told which node has
// which children. I do most of my debugging/reconstruction by hand-drawing the
// output of this function.
func printTree(w *walletNode) {
	if w == nil {
		return
	}

	fmt.Println(w)
	printTree(w.children[0])
	printTree(w.children[1])
}

// findRedViolatoins iterates through the rbw tree starting with w as the root
// and verifies that in no situations does a red node have another red node as
// a child.
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

// findBlackViolations iterates through rbw tree starting with w as the root
// and makes sure that the height of black nodes in the tree is the same at all
// places in the tree.
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

// findSortViolaitons iterates through the rbw tree starting with w as the root
// and makes sure that the entire tree is sorted with no out-of-order elements.
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

// checkViolations goes through a tree and verifies it as a proper Sia
// red-black tree. This means checking for red violations, black violations,
// sort violations, and then makes sure that all of the weights in the tree
// make sense. It does not count the number of reachable nodes because it
// doesn't know how many nodes are supposed to be reachable. checkViolations
// takes the testing.T object as input so it can call t.Error directly.
func checkViolations(location int, w *walletNode, expectedReachable int, t *testing.T) {
	redViolations := findRedViolations(w)
	if redViolations != 0 {
		t.Error(location, "- Got red violations:", redViolations)
	}

	_, blackViolations := findBlackViolations(w)
	if blackViolations != 0 {
		t.Error(location, "- Got black violations:", blackViolations)
	}

	sortViolations := findSortViolations(w)
	if sortViolations != 0 {
		t.Error(location, "- Got sort violations:", sortViolations)
	}

	weightViolations := findWeightViolations(w)
	if weightViolations != 0 {
		t.Error(location, "- Got weight violations:", weightViolations)
	}

	reachableNodes := countReachableNodes(w)
	if reachableNodes != expectedReachable {
		t.Error(location, "- Wrong number of reachable nodes:", reachableNodes)
	}
}

// createTestWallet produces a wallet where the id and the weight are both set
// to the same number, which is given in the input.
func createTestWalletNode(id int) (w *walletNode) {
	w = new(walletNode)
	w.id = WalletID(id)
	w.weight = id
	return
}

// Extensive testing to make sure that there are no errors with the red-black
// tree. This test is particularly fleshed out because red-black trees are
// tricky and I made many mistakes in the first implementation.
func TestWalletTree(t *testing.T) {
	// create a quorum and add a few children, just enough to hit a tree height
	// of 4. After each insertion the tree is verified for integrity.
	q := new(Quorum)
	n := 9
	for i := 0; i < n; i++ {
		newWallet := createTestWalletNode(i)
		q.insert(newWallet)
		checkViolations(0, q.walletRoot, i+1, t)
	}

	// make sure that all inserted nodes are reachable
	for i := 0; i < n; i++ {
		w := q.retrieve(WalletID(i))
		if w == nil {
			t.Error("Unreachable Node:", i)
		}
	}

	// remove each of the inserted nodes. After each removal, the tree is
	// verified for integrity.
	for i := 0; i < n; i++ {
		q.remove(WalletID(i))

		w := q.retrieve(WalletID(i))
		if w != nil {
			t.Error("Maganed to retrieve a removed wallet:", i)
		}
		checkViolations(1, q.walletRoot, n-1-i, t)
	}

	if testing.Short() {
		t.Skip()
	}

	// The following tests randomly insert and delete nodes in mass in hopes of
	// hitting every possible test case.

	// Insert many nodes into the tree, each with a random weight. After each
	// iteration, verify the integrity of the tree.
	for z := 0; z < 50; z++ {
		n = 257
		weights := make(map[uint64]bool) // keeps track of which elements have been added
		for i := 0; i < n; i++ {
			found := false
			var weight int
			var err error
			for !found {
				weight, err = siacrypto.RandomInt(100000)
				if err != nil {
					t.Fatal(err)
				}

				if weights[uint64(weight)] == false {
					weights[uint64(weight)] = true
					found = true
				}
			}

			node := createTestWalletNode(weight)
			q.insert(node)
			checkViolations(2, q.walletRoot, i+1, t)
		}

		// randomly choose between inserting and deleting a random item
		for i := 0; i < n; i++ {
			insertOrDelete, err := siacrypto.RandomInt(2) // [0, 2)
			if err != nil {
				t.Fatal(err)
			}

			if insertOrDelete == 0 {
				// insert
				found := false
				var weight int
				var err error
				for !found {
					weight, err = siacrypto.RandomInt(100000)
					if err != nil {
						t.Fatal(err)
					}

					if weights[uint64(weight)] == false {
						weights[uint64(weight)] = true
						found = true
					}
				}

				node := createTestWalletNode(weight)
				q.insert(node)
				checkViolations(4, q.walletRoot, len(weights), t)
			} else {
				// delete
				// turn weights into a slice so that a value can be selected at random
				j := 0
				weightSlice := make([]uint64, len(weights))
				for key := range weights {
					weightSlice[j] = key
					j++
				}

				j, err = siacrypto.RandomInt(len(weightSlice))
				if err != nil {
					t.Fatal(err)
				}

				q.remove(WalletID(weightSlice[j]))
				delete(weights, weightSlice[j])
				checkViolations(5, q.walletRoot, len(weights), t)
			}
		}

		// verify that all weights within the slice are reachable
		for key := range weights {
			w := q.retrieve(WalletID(key))
			if w == nil {
				t.Error("Retrieval Error:", key)
			}
		}

		// delete all of the remaining items in the map in a random order
		i := 0
		weightSlice := make([]uint64, len(weights))
		for key := range weights {
			weightSlice[i] = key
			i++
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
			checkViolations(3, q.walletRoot, len(weightSlice)-1-i, t)
		}
	}
}
