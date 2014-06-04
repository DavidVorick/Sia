package quorum

import (
	"os"
)

// A wallet node is the base unit for the WalletTree. The wallet tree is a
// red-black tree sorted by id. It's used to load balance between quorums and
// to pick random sectors in logarithmic time. It's also currently used for
// lookup, but lookup may be moved to a map such that lookup happens in linear
// time.
type walletNode struct {
	red      bool
	children [2]*walletNode

	id     WalletID
	weight int
	wallet os.File
}

// not Prevents redundant code for symetrical cases. Theres a direction, and then
// there's the opposite of a direction. This function returns the opposite of
// that direction.
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

// Does a rotation within the red-black tree, while keeping all weights correct.
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
	tmpWeight := w.weight
	w.weight = 0

	// search down the tree
	for {
		if current == nil {
			// insert new node at bottom
			parent.children[direction] = w
			current = w
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

		// stop at bottom
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
		i := falseRoot
		targetWeight := target.weight
		if target.children[0] != nil {
			targetWeight -= target.children[0].weight
		}
		if target.children[1] != nil {
			targetWeight -= target.children[1].weight
		}

		for i != nil && i != target {
			i.weight -= targetWeight

			if i.id > current.id {
				i = i.children[0]
			} else {
				i = i.children[1]
			}
		}
		for i != nil && i != current {
			i.weight -= current.weight

			if i.id > current.id {
				i = i.children[0]
			} else {
				i = i.children[1]
			}
		}

		direction0 := 0
		if parent.children[1] == current {
			direction0 = 1
		}
		parent.children[direction0] = current.children[0]

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
	q.walletRoot = falseRoot.children[1]
	if q.walletRoot != nil {
		q.walletRoot.red = false
	}

	return
}

func (q *Quorum) retrieve(id WalletID) *walletNode {
	current := q.walletRoot

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
