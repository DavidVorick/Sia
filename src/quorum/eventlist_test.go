package quorum

/* import (
	"siacrypto"
	"testing"
)

func countReachableEvents(en *eventNode) (i int) {
	for {
		if en == nil {
			return
		}
		i += 1

		current := en.top
		if current == nil {
			return
		}
		current = current.bottom()
		if current.nextNode == en {
			panic("countReachableEvents has detected an infinite loop")
		}
		en = current.nextNode
	}
}

func TestEventList(t *testing.T) {
	u0 := &upload{
		deadline: 5,
	}

	q := new(Quorum)
	q.insertEvent(u0)

	en0 := q.eventNode(u0)
	if en0 == nil {
		t.Fatal("Could not get inserted event!")
	}
	if countReachableEvents(q.eventRoot) != 1 {
		t.Fatal("Reached wrong number of events, expecting 1:", countReachableEvents(q.eventRoot))
	}

	q.deleteEvent(u0)
	en0 = q.eventNode(u0)
	if en0 != nil {
		t.Fatal("deleted event node still retrievable")
	}
	if countReachableEvents(q.eventRoot) != 0 {
		t.Fatal("Reached wrong number of events, expecting 0:", countReachableEvents(q.eventRoot))
	}

	if testing.Short() {
		t.Skip()
	}

	uploadMap := make(map[*upload]*upload)
	n := 100
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			randomTimeout, _ := siacrypto.RandomInt(12)
			nu := &upload{
				deadline: uint32(randomTimeout),
			}
			uploadMap[nu] = nu
			q.insertEvent(nu)

			if countReachableEvents(q.eventRoot) != i+1 {
				t.Error("Reached wrong number of events, expecting", i+1, "got", countReachableEvents(q.eventRoot))
			}
		}

		elementSlice := make([]*upload, n)
		i := 0
		for key := range uploadMap {
			elementSlice[i] = key
			i++
		}

		// try and fetch every element
		for i := range elementSlice {
			wn := q.eventNode(elementSlice[i])
			if wn == nil {
				t.Error("cannot reach inserted element")
			}
		}

		// shuffle elementSlice
		for i := range elementSlice {
			newIndex, err := siacrypto.RandomInt(len(elementSlice) - i)
			if err != nil {
				t.Fatal(err)
			}
			newIndex += i

			tmp := elementSlice[newIndex]
			elementSlice[newIndex] = elementSlice[i]
			elementSlice[i] = tmp
		}

		for i := range elementSlice {
			q.deleteEvent(elementSlice[i])
			wn := q.eventNode(elementSlice[i])
			if wn != nil {
				t.Error("deleted event node is still fetchable")
			}
			if countReachableEvents(q.eventRoot) != n-i-1 {
				t.Fatal("Wrong number of reachable events, expecting", n-i-1, "got", countReachableEvents(q.eventRoot))
			}
		}
		uploadMap = make(map[*upload]*upload)
	}

	// insert a bunch of random things
	// randomly insert and delete the things
	// delete all of the things
	// check sorting each time
} */
