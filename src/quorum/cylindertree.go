package quorum

// cylinderNode is the basic element in the AA tree that enables efficient
// selection of random sectors for verification
type cylinderNode struct {
	parent *cylinderNode
	left   *cylinderNode
	right  *cylinderNode

	children int
	weight   int

	data *Cylinder
}

// insert takes a cylinder object that is not yet in the cylinderTree and puts it
// into the cylinderTree
func (q *quorum) insert(cn *cylinderNode) {
	// insert the node into the lightest-weight half of the parent
	currentNode := q.cylinderTreeHead
	for currentNode.children > 1 {
		if currentNode.left.children < currentNode.right.children {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}
	}
	// either one child is nil (max 1 child) or both are nil
	if currentNode.left == nil {
		currentNode.left = cn
	} else {
		currentNode.right = cn
	}
	cn.parent = currentNode

	// update the aggregate values of all parents
	for currentNode != nil {
		currentNode.children += 1
		currentNode.weight += cn.weight
		currentNode = currentNode.parent
	}
}

// delete takes a node from the cylinderTree and deletes it from the tree
func (q *quorum) delete(bn *cylinderNode) {
	// get a replacement node from the heaviest part of the tree, removing it
	var replacementNode *cylinderNode
	currentNode := q.cylinderTreeHead
	for currentNode.children > 2 {
		if currentNode.left.children > currentNode.right.children {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}
	}
	// Our parent has at least 3 children, the heaviest side was chosen, meaning
	// the current node has at least 1 child
	if currentNode.left == nil {
		replacementNode = currentNode.right
		currentNode.right = nil
	} else {
		replacementNode = currentNode.left
		currentNode.left = nil
	}
	for currentNode != nil {
		currentNode.children -= 1
		currentNode.weight -= replacementNode.weight
		currentNode = currentNode.parent
	}

	// nothing needs to be done if bn was already the heaviest node
	if replacementNode != bn {
		// place replacementNode as the child of the parent
		currentNode = bn.parent
		if currentNode.left == bn {
			currentNode.left = replacementNode
		} else {
			currentNode.right = replacementNode
		}
		replacementNode.parent = currentNode

		// calculate weight of bn
		if bn.left != nil {
			bn.weight -= bn.left.weight
		}
		if bn.right != nil {
			bn.weight -= bn.right.weight
		}

		// update the weights of the parents
		for currentNode != nil {
			currentNode.weight -= bn.weight
			currentNode.weight += replacementNode.weight
			currentNode = currentNode.parent
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
func (parent *cylinderNode) insertDelete(insertBN, deleteBN *cylinderNode) {
	// tbi
}

// randomSector takes a random int between 0 and the total weight of the
// cylinderTree and picks a sector at random to be used in the proof-of-storage
func (q *quorum) randomSector() (c *Cylinder, sector int) {
	// get a random number between 0 and the batch tree weight
	random, err := q.randInt(0, q.cylinderTreeHead.weight)
	if err != nil {
		// not sure
		return
	}

	// tree is post-ordered; the parent comes after the children
	// this just makes the code a bit cleaner
	currentNode := q.cylinderTreeHead
	for {
		// check for nil statemenets
		if currentNode.left == nil && currentNode.right == nil {
			break
		} else if currentNode.left == nil {
			if random < currentNode.right.weight {
				currentNode = currentNode.right
			}
			break
		} else if currentNode.right == nil {
			if random < currentNode.left.weight {
				currentNode = currentNode.left
			} else {
				random -= currentNode.left.weight
			}
			break
		}

		// logic if no nil statements are found
		if random < currentNode.left.weight {
			currentNode = currentNode.left
		} else if random < (currentNode.left.weight + currentNode.right.weight) {
			random -= currentNode.left.weight
			currentNode = currentNode.right
		} else {
			random -= currentNode.left.weight
			random -= currentNode.right.weight
			break
		}
	}
	c = currentNode.data

	// figure out which index to use
	for index, value := range currentNode.data.RingAtoms {
		if value > random {
			sector = index
			break
		}
		random -= value
	}
	return
}
