package quorum

import (
	"testing"
)

func TestBatchTree(t *testing.T) {
	// create a parent node
	parent := new(batchNode)
	parent.weight = 1

	child0 := new(batchNode)
	child0.weight = 2
	child1 := new(batchNode)
	child1.weight = 3
	child2 := new(batchNode)
	child2.weight = 4

	parent.insert(child0)
	parent.insert(child1)
	parent.insert(child2)

	if parent.weight != 10 {
		t.Error("parent.weight not updating correctly with insert:", parent.weight)
	}
	if parent.children != 3 {
		t.Error("parent.children not updating correctly with insert:", parent.children)
	}

	println("starting 0")
	parent.delete(child0)
	println("starting 1")
	parent.delete(child1)
	println("starting 2")
	parent.delete(child2)
	println("finished")

	if parent.weight != 1 {
		t.Error("parent.weight not updating correctly with delete:", parent.weight)
	}
	if parent.children != 0 {
		t.Error("parent.children not updating correctly with delete:", parent.children)
	}

	// fuzzing, add a bunch of elements of random weights, each time checking the
	// whole tree for integrity. then randomly add or delete random elements each
	// time checking the whole tree for integrity. Finally, remove all remaining
	// elements each time checking the whole tree for integrity.
	//
	// a separate data structure will be needed to know how much weight each
	// element is supposed to have
}
