package quorum

import (
	"common"
	"common/crypto"
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

// Adds a new Sibling, and then announces them with their index
// Currently not safe - Siblings need to be added during compile()
func (p *Participant) JoinSia(s Sibling, arb *struct{}) (err error) {
	p.broadcast(&common.Message{
		Proc: "Participant.AddHopeful",
		Args: s,
		Resp: nil,
	})
	return
}

// add a potential sibling to the heartbeat-in-progress
func (p *Participant) AddHopeful(s Sibling, arb *struct{}) (err error) {
	fmt.Println("got join request")
	p.currHeartbeatLock.Lock()
	p.currHeartbeat.hopefuls = append(p.currHeartbeat.hopefuls, &s)
	p.currHeartbeatLock.Unlock()
	return
}

func (p *Participant) TransferQuorum(encodedQuorum []byte, arb *struct{}) (err error) {
	err = p.quorum.GobDecode(encodedQuorum)
	fmt.Println("downloaded quorum:")
	fmt.Print(p.quorum.Status())
	// determine our index by searching through the quorum
	// also create maps for each sibling
	for i, s := range p.quorum.siblings {
		if s.compare(p.self) {
			p.self.index = byte(i)
			p.addNewSibling(p.self)
		} else {
			p.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
		}
	}
	p.tickingLock.Lock()
	defer p.tickingLock.Unlock()
	if !p.ticking {
		p.ticking = true
		go p.tick()
	}
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
