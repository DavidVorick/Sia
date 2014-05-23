package quorum

import (
	"bytes"
	"common"
	"common/crypto"
	"encoding/gob"
	"fmt"
)

/*
The Bootstrapping Process
1. The new sibling announces its intent to the quorum.
2. The quorum includes the sibling as a "hopeful" in the next heartbeat.
3. During compile, the quorum decides whether or not to add the hopeful, and where.
4. If accepted, the hopeful downloads the current quorum state.
5. The current quorum siblings add the new participant, along with the default heartbeat.
6. The hopeful listens to the quorum and processes any incoming heartbeats.
7. After the next compile, the hopeful becomes a full sibling.


[- Interim 0 -]       [-- Compile --]       [- Interim 1 -]       [-- Compile --]       [- Interim 2 -]       [-- Compile --]       [- Interim 3 -]       [-- Compile --]
[   hopeful   ]       [             ]       [   hopeful   ]       [   quorum    ]       [ hopeful gets]       [ default hb  ]       [   hopeful   ]       [             ]
[  announces  ]  -->  [             ]  -->  [  added to   ]  -->  [ decides to  ]  -->  [  state and  ]  -->  [  used for   ]  -->  [  now fully  ]  -->  [             ]
[   intent    ]       [             ]       [  heartbeat  ]       [ add hopeful ]       [  heartbeats ]       [   compile   ]       [  integrated ]       [             ]
[-------------]       [-------------]       [-------------]       [-------------]       [-------------]       [-------------]       [-------------]       [-------------]

*/

// Bootstrapping
var bootstrapAddress = common.Address{
	ID:   1,
	Host: "localhost",
	Port: 9988,
}

// A Join is an update that is a participant requesting to join Sia.
type Join struct {
	Sibling   Sibling
	Heartbeat SignedHeartbeat
}

func (j *Join) process(p *Participant) {
	// add hopefuls to any available slots
	// quorum is already locked by compile()
	i := 0
	for i < common.QuorumSize {
		if p.quorum.siblings[i] == nil {
			// transfer the quorum to the new sibling
			go func() {
				// wait until compile() releases the mutex
				p.quorum.lock.RLock()
				gobQuorum, _ := p.quorum.GobEncode()
				p.quorum.lock.RUnlock() // quorum can be unlocked as soon as GobEncode() completes
				p.messageRouter.SendAsyncMessage(&common.Message{
					Dest: j.Sibling.address,
					Proc: "Participant.TransferQuorum",
					Args: gobQuorum,
					Resp: nil,
				})
			}()
			j.Sibling.index = byte(i)
			p.addNewSibling(&j.Sibling)
			println("placed hopeful at index", i)
			break
		}
		i++
	}
}

func (j *Join) GobEncode() (gobJoin []byte, err error) {
	if j == nil {
		err = fmt.Errorf("Cannot encode a nil Join")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(j.Sibling)
	if err != nil {
		return
	}
	err = encoder.Encode(j.Heartbeat)
	if err != nil {
		return
	}
	gobJoin = w.Bytes()
	return
}

func (j *Join) GobDecode(gobJoin []byte) (err error) {
	if j == nil {
		j = new(Join)
	}

	r := bytes.NewBuffer(gobJoin)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&j.Sibling)
	if err != nil {
		return
	}
	err = decoder.Decode(&j.Heartbeat)
	if err != nil {
		return
	}
	return
}

func (p *Participant) JoinSia(s Sibling, arb *struct{}) (err error) {
	p.broadcast(&common.Message{
		Proc: "Participant.AddHopeful",
		Args: s,
		Resp: nil,
	})
	return
}

// A member of a quorum will call TransferQuorum on someone who has solicited
// a quorum transfer. The quorum data is then sent between machines over RPC.
func (p *Participant) TransferQuorum(encodedQuorum []byte, arb *struct{}) (err error) {
	// lock the quorum before making major changes
	p.quorum.lock.Lock()
	err = p.quorum.GobDecode(encodedQuorum)
	p.quorum.lock.Unlock()

	fmt.Println("downloaded quorum:")
	fmt.Print(p.quorum.Status())

	// determine our index by searching through the quorum
	// also create maps for each sibling
	p.quorum.lock.RLock()
	for i, s := range p.quorum.siblings {
		if s.compare(p.self) {
			p.self.index = byte(i)
			p.addNewSibling(p.self)
		} else {
			p.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
		}
	}
	p.quorum.lock.RUnlock()
	go p.tick()
	return
}

// Add a Sibling to the state, tell the Sibling about ourselves
// Note: p.heartbeats and p.quorum.siblings are already locked by compile()
func (p *Participant) addNewSibling(s *Sibling) (err error) {
	// make the heartbeat map and add the default heartbeat
	hb := new(heartbeat)

	// get the hash of the default heartbeat
	ehb, err := hb.GobEncode()
	if err != nil {
		return
	}
	hbHash, err := crypto.CalculateTruncatedHash(ehb)
	if err != nil {
		return
	}

	p.heartbeats[s.index] = make(map[crypto.TruncatedHash]*heartbeat)
	p.heartbeats[s.index][hbHash] = hb
	p.quorum.siblings[s.index] = s

	return
}
