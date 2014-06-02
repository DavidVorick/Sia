package quorum

type WalletID uint64

// wallet is the basic element in the AA tree that enables efficient
// selection of random sectors for verification
type wallet struct {
	id WalletID

	parent *wallet
	left   *wallet
	right  *wallet

	weight   int
	children int
}

// insert takes a wallet object that is not yet in the walletTree and puts it
// into the walletTree
func (q *Quorum) insert(w *wallet) {
	// insert the node into the lightest-weight half of the parent
	current := q.walletTreeHead
	for current.children > 1 {
		if current.left.children < current.right.children {
			current = current.left
		} else {
			current = current.right
		}
	}
	// either one child is nil (max 1 child) or both are nil
	if current.left == nil {
		current.left = w
	} else {
		current.right = w
	}
	w.parent = current

	// update the aggregate values of all parents
	for current != nil {
		current.children += 1
		current.weight += w.weight
		current = current.parent
	}
}

// delete takes a node from the walletTree and deletes it from the tree
func (q *Quorum) delete(bn *wallet) {
	// get a replacement node from the heaviest part of the tree, removing it
	var replacementNode *wallet
	current := q.walletTreeHead
	for current.children > 2 {
		if current.left.children > current.right.children {
			current = current.left
		} else {
			current = current.right
		}
	}
	// Our parent has at least 3 children, the heaviest side was chosen, meaning
	// the current node has at least 1 child
	if current.left == nil {
		replacementNode = current.right
		current.right = nil
	} else {
		replacementNode = current.left
		current.left = nil
	}
	for current != nil {
		current.children -= 1
		current.weight -= replacementNode.weight
		current = current.parent
	}

	// nothing needs to be done if bn was already the heaviest node
	if replacementNode != bn {
		// place replacementNode as the child of the parent
		current = bn.parent
		if current.left == bn {
			current.left = replacementNode
		} else {
			current.right = replacementNode
		}
		replacementNode.parent = current

		// calculate weight of bn
		if bn.left != nil {
			bn.weight -= bn.left.weight
		}
		if bn.right != nil {
			bn.weight -= bn.right.weight
		}

		// update the weights of the parents
		for current != nil {
			current.weight -= bn.weight
			current.weight += replacementNode.weight
			current = current.parent
		}

		// update weights and pointers
		replacementNode.left = bn.left
		replacementNode.right = bn.right
		replacementNode.children = bn.children
		bn.left = nil
		bn.right = nil
		if replacementNode.left != nil {
			replacementNode.weight += replacementNode.left.weight
		}
		if replacementNode.right != nil {
			replacementNode.weight += replacementNode.right.weight
		}
	}
}

// insertDelete takes an element to be inserted and an element to be deleted and
// performs both operations at once. Doing both operations at the same time
// means that less work overall must be performed; you replace the deleted
// element with the inserted element, and then you update the parent-set once.
// This is less work than even a single insert or a single delete.
func (parent *wallet) insertDelete(insertBN, deleteBN *wallet) {
	// tbi
}

// randomSector takes a random int between 0 and the total weight of the
// walletTree and picks a sector at random to be used in the proof-of-storage
func (q *Quorum) randomSector() (c *wallet) {
	// get a random number between 0 and the batch tree weight
	random, err := q.randInt(0, q.walletTreeHead.weight)
	if err != nil {
		// not sure
		return
	}

	// tree is post-ordered; the parent comes after the children
	// this just makes the code a bit cleaner
	current := q.walletTreeHead
	for {
		// check for nil statemenets
		if current.left == nil && current.right == nil {
			break
		} else if current.left == nil {
			if random < current.right.weight {
				current = current.right
			}
			break
		} else if current.right == nil {
			if random < current.left.weight {
				current = current.left
			} else {
				random -= current.left.weight
			}
			break
		}

		// logic if no nil statements are found
		if random < current.left.weight {
			current = current.left
		} else if random < (current.left.weight + current.right.weight) {
			random -= current.left.weight
			current = current.right
		} else {
			random -= current.left.weight
			random -= current.right.weight
			break
		}
	}
	return
}
