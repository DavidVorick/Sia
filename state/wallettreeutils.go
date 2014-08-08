package state

import (
	"fmt"
)

// wallettreeutils.go contains functions that interact with the wallet tree but
// aren't actually a part of the weighted-red-black tree data structure. I felt
// that they should be in a separate file because the weighted-red-black tree
// logic is complex enough to merit it's own file.

func (wn *walletNode) nodeWeight() (nw int) {
	if wn == nil {
		return
	}
	nw = wn.weight
	if wn.children[0] != nil {
		nw -= wn.children[0].weight
	}
	if wn.children[1] != nil {
		nw -= wn.children[1].weight
	}
	return
}

func (s *State) updateWeight(id WalletID, delta int) (err error) {
	// check that the id is in the quorum
	wn := s.walletNode(id)
	if wn == nil {
		err = fmt.Errorf("id not found in wallet tree")
		return
	}

	if s.walletRoot.weight+delta > AtomsPerQuorum {
		err = fmt.Errorf("Insufficient room in quorum to complete action")
		return
	}

	currentNode := s.walletRoot
	for currentNode.id != id {
		currentNode.weight += delta
		if currentNode.id > id {
			currentNode = currentNode.children[0]
		} else {
			currentNode = currentNode.children[1]
		}
	}
	currentNode.weight += delta
	return
}

// buildWalletList is a helper function that does a recursive DFS through the
// wallet tree, recording the ids of every wallet as it progresses.
// buildWalletList uses all pointers, and acts on the underlaying objects of
// its inputs, which is why it doesn't return anything.
func buildWalletList(w *walletNode, wd []WalletID, index *int) {
	if w == nil {
		return
	}

	buildWalletList(w.children[0], wd, index)

	wd[*index] = w.id
	*index++

	buildWalletList(w.children[1], wd, index)
}

// WalletList returns a list of every wallet that can be found in the wallet
// tree, sorted by id.
func (s *State) WalletList() (wd []WalletID) {
	wd = make([]WalletID, s.wallets)
	initialIndex := 0
	buildWalletList(s.walletRoot, wd, &initialIndex)
	return
}
