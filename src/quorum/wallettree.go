package quorum

type walletNode struct {
	red      bool
	children [2]*walletNode

	id     WalletID
	weight int
}

func not(direction int) int {
	if direction == 0 {
		return 1
	}
	return 0
}

func (w *walletNode) isRed() bool {
	return w != nil && w.red
}

func (w *walletNode) rotate(direction int) *walletNode {
	tmp := w.children[not(direction)]
	w.children[not(direction)] = tmp.children[direction]
	tmp.children[direction] = w
	w.red = true
	tmp.red = false
	return tmp
}

func (w *walletNode) doubleRotate(direction int) *walletNode {
	w.children[not(direction)] = w.children[not(direction)].rotate(not(direction))
	return w.rotate(direction)
}

func (q *Quorum) insert(w *walletNode) {
	if w == nil {
		return
	}
	w.red = true

	if q.walletRoot == nil {
		q.walletRoot = w
		q.walletRoot.red = false
		return
	}

	// helper variables
	falseRoot := new(walletNode)
	var grandparent *walletNode
	var parent *walletNode
	temp := falseRoot
	current := q.walletRoot
	temp.children[1] = q.walletRoot
	direction := 0
	previousDirection := 0

	// search down the tree
	for {
		if current == nil {
			// insert new node at bottom
			parent.children[direction] = w
			current = w
			if current == nil {
				return
			}
		} else if current.children[0].isRed() && current.children[1].isRed() {
			// color flip
			current.red = true
			current.children[0].red = false
			current.children[1].red = false
		}

		// Fix red violation
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

		// stop if wallet already exists
		if current.id == w.id {
			break
		}

		previousDirection = direction
		direction = 0
		if current.id < w.id {
			direction = 1
		}

		// update helpers
		if grandparent != nil {
			temp = grandparent
		}
		grandparent = parent
		parent = current
		current = current.children[direction]
	}

	q.walletRoot = falseRoot.children[1]
	q.walletRoot.red = false
}

func (q *Quorum) remove(id WalletID) (target *walletNode) {
	if q.walletRoot == nil {
		return
	}

	// helpers
	falseRoot := new(walletNode)
	var grandparent *walletNode
	var parent *walletNode
	current := falseRoot
	current.children[1] = q.walletRoot
	direction := 1

	// search and push down a red
	for current.children[direction] != nil {
		previousDirection := direction

		// update helpers
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
		target = current

		direction0 := 0
		direction1 := 0
		if parent.children[1] == current {
			direction0 = 1
		}
		if current.children[0] == nil {
			direction1 = 1
		}
		parent.children[direction0] = current.children[direction1]
	}

	// update root and make it black
	q.walletRoot = falseRoot.children[1]
	if q.walletRoot != nil {
		q.walletRoot.red = false
	}

	return
}
