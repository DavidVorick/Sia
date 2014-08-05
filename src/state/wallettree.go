package state

// A walletNode is the base unit for the WalletTree. The wallet tree is a
// red-black tree sorted by id. It's used to load balance between quorums and
// to pick random sectors in logarithmic time. It's also currently used for
// lookup, but lookup may be moved to a map such that lookup happens in linear
// time.
//
// walletNode composes a rbw tree, which is a red-black-weighted tree. Each
// node has a weight, and that weight is it's internal weight plus all of the
// internal weights of all of its children. These weights are used to pick a
// random sector in log(n) time. The tree is sorted by the id.
//
// A huge thanks goes to the page at EternallyConfuzzled explaining red-black
// trees. The influence of that page is obvious within this tree.
type walletNode struct {
	red      bool
	children [2]*walletNode

	id     WalletID
	weight int
}

// not is a helper function that emulates C's '!' for integers.
// It prevents redundant code for symmetrical cases.
func not(direction int) int {
	if direction == 0 {
		return 1
	}
	return 0
}

// isRed returns true if a node is red and false if a node is black or is nil
func (w *walletNode) isRed() bool {
	return w != nil && w.red
}

// rotate performs a rotation within the red-black tree, while keeping all weights correct.
func (w *walletNode) rotate(direction int) *walletNode {
	if w.children[not(direction)] != nil {
		w.weight -= w.children[not(direction)].weight
		w.children[not(direction)].weight += w.weight
		if w.children[not(direction)].children[direction] != nil {
			w.weight += w.children[not(direction)].children[direction].weight
		}
	}

	tmp := w.children[not(direction)]
	w.children[not(direction)] = tmp.children[direction]
	tmp.children[direction] = w
	w.red = true
	tmp.red = false
	return tmp
}

// doubleRotate performs a double rotation within the rbw tree
func (w *walletNode) doubleRotate(direction int) *walletNode {
	w.children[not(direction)] = w.children[not(direction)].rotate(not(direction))
	return w.rotate(direction)
}

// insert takes a walletNode and inserts it into the rbw tree held within the quorum.
func (s *State) insertWalletNode(w *walletNode) {
	// exit insertion if given a nil node to insert
	if w == nil {
		return
	}
	w.red = true // all nodes are inserted as red

	// if the root is nil, insert the node at the root and make it black.
	if s.walletRoot == nil {
		s.walletRoot = w
		s.walletRoot.red = false
		s.wallets++
		return
	}

	// helper variables
	falseRoot := new(walletNode)
	var grandparent *walletNode
	var parent *walletNode
	temp := falseRoot
	current := s.walletRoot
	temp.children[1] = s.walletRoot
	direction := 0
	previousDirection := 0

	// nodes are inserted as having 0 weight to cause the least disruption
	// possible
	tmpWeight := w.weight
	w.weight = 0

	// iterate through the tree, looking for an insertion location
	for {
		if current == nil {
			// insert new node at bottom
			parent.children[direction] = w
			current = w
		} else if current.children[0].isRed() && current.children[1].isRed() {
			// color flip if both children are red
			current.red = true
			current.children[0].red = false
			current.children[1].red = false
		}

		// insertion and/or colorflipping may cause a red violation, this corrects
		// that violation
		if current.isRed() && parent.isRed() {
			direction2 := 0
			if temp.children[1] == grandparent {
				direction2 = 1
			}

			// single rotate in one case, double rotate in the other
			if current == parent.children[previousDirection] {
				temp.children[direction2] = grandparent.rotate(not(previousDirection))
			} else {
				temp.children[direction2] = grandparent.doubleRotate(not(previousDirection))
			}
		}

		// stop if we have reached the node that we inserted
		if current.id == w.id {
			break
		}

		// pick the next direction
		previousDirection = direction
		direction = 0
		if current.id < w.id {
			direction = 1
		}

		// move the helpers forward one generation
		if grandparent != nil {
			temp = grandparent
		}
		grandparent = parent
		parent = current
		current = current.children[direction]
	}

	// after insertion, restore the inserted node to its original weight. Then
	// iterate through it's parents (starting from the root and working down) and
	// add that weight to all of their weight representations.
	w.weight += tmpWeight
	i := falseRoot.children[1]
	for i != nil && i != w {
		i.weight += tmpWeight

		if i.id > current.id {
			i = i.children[0]
		} else {
			i = i.children[1]
		}
	}

	// restore the root wallet and set it to black
	s.walletRoot = falseRoot.children[1]
	s.walletRoot.red = false
	s.wallets++
}

// remove removes the presented key from the wallet tree.
func (s *State) removeWalletNode(id WalletID) (target *walletNode) {
	// if the tree is nil, there is nothing to do
	if s.walletRoot == nil {
		return
	}

	// initialize helper variables
	falseRoot := new(walletNode)
	var parent, grandparent *walletNode
	current := falseRoot
	current.children[1] = s.walletRoot
	direction := 1

	// search and push down a red node
	for current.children[direction] != nil {
		previousDirection := direction

		// advance the helpers a generation
		grandparent = parent
		parent = current
		current = current.children[direction]
		direction = 0
		if current.id < id {
			direction = 1
		}

		// check if current is the desired wallet
		if current.id == id {
			target = current
		}

		// push red node down
		if !current.isRed() && !current.children[direction].isRed() {
			if current.children[not(direction)].isRed() {
				parent.children[previousDirection] = current.rotate(direction)
				parent = parent.children[previousDirection]
			} else if !current.children[not(direction)].isRed() {
				temp := parent.children[not(previousDirection)]

				if temp != nil {
					if !temp.children[not(previousDirection)].isRed() && !temp.children[previousDirection].isRed() {
						// color flip
						parent.red = false
						temp.red = true
						current.red = true
					} else {
						direction2 := 0
						if grandparent.children[1] == parent {
							direction2 = 1
						}

						if temp.children[previousDirection].isRed() {
							grandparent.children[direction2] = parent.doubleRotate(previousDirection)
						} else if temp.children[not(previousDirection)].isRed() {
							grandparent.children[direction2] = parent.rotate(previousDirection)
						}

						// ensure correct coloring
						grandparent.children[direction2].red = true
						current.red = true
						grandparent.children[direction2].children[0].red = false
						grandparent.children[direction2].children[1].red = false
					}
				}
			}
		}
	}

	// replace and remove if found
	if target != nil {
		// figure out what the original weight of the target node is.
		i := falseRoot
		targetWeight := target.weight
		if target.children[0] != nil {
			targetWeight -= target.children[0].weight
		}
		if target.children[1] != nil {
			targetWeight -= target.children[1].weight
		}

		// iterate through every node from the root to the target, subtracting the
		// target's weight.
		for i != nil && i != target {
			i.weight -= targetWeight

			if i.id > current.id {
				i = i.children[0]
			} else {
				i = i.children[1]
			}
		}

		// iterate through every node from the target to the current, subtracting
		// the current's weight.
		for i != nil && i != current {
			i.weight -= current.weight

			if i.id > current.id {
				i = i.children[0]
			} else {
				i = i.children[1]
			}
		}

		// remove the current node from the rbw tree
		direction0 := 0
		if parent.children[1] == current {
			direction0 = 1
		}
		parent.children[direction0] = current.children[0]

		// mark the current node as the new target, updating the weights and data
		// to reflect the changed position.
		target.id = current.id
		target.weight = current.weight
		if target.children[0] != nil {
			target.weight += target.children[0].weight
		}
		if target.children[1] != nil {
			target.weight += target.children[1].weight
		}
	}

	// update root and make it black
	s.walletRoot = falseRoot.children[1]
	if s.walletRoot != nil {
		s.walletRoot.red = false
	}
	s.wallets--

	return
}

// Fetches the wallet from the rbw tree that matches the id presented.
func (s *State) walletNode(id WalletID) *walletNode {
	current := s.walletRoot

	for current != nil {
		if current.id == id {
			return current
		}

		if current.id > id {
			current = current.children[0]
		} else {
			current = current.children[1]
		}
	}

	return nil
}
