package quorum

// batchNode is the basic element in the AA tree that enables efficient
// selection of random sectors for verification
type batchNode struct {
	parent *batchNode
	left   *batchNode
	right  *batchNode

	children int
	weight   int
}

// insert takes a batch object that is not yet in the batchTree and puts it
// into the batchTree
func (parent *batchNode) insert(batch *batch) {
	batch.node = new(batchNode)
	for _, value := range batch.sectorLengths {
		batch.node.weight += value
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
	// either one child is nil (max 1 child) or both are nil
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

// delete takes a node from the batchTree and deletes it from the tree
func (parent *batchNode) delete(bn *batchNode) {
	// get a replacement node from the heaviest part of the tree, removing it
	var replacementNode *batchNode
	currentNode := parent
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

		// calculate weight of bn
		bnWeight := bn.weight
		if bn.left != nil {
			bnWeight -= bn.left.weight
		}
		if bn.right != nil {
			bnWeight -= bn.right.weight
		}

		// update the weights of the parents
		for currentNode != nil {
			currentNode.weight -= bnWeight
			currentNode.weight += replacementNode.weight
		}

		// update replacementNode weight and pointers
		replacementNode.left = bn.left
		replacementNode.right = bn.right
		if replacementNode.left != nil {
			replacementNode.weight += replacementNode.left.weight
		}
		if replacementNode.right != nil {
			replacementNode.weight += replacementNode.right.weight
		}
	}
}
