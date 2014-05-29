package quorum

import (
	"bytes"
	"common"
	"encoding/gob"
	"fmt"
	"siacrypto"
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
type JoinRequest struct {
	Sibling Sibling
}

func (j JoinRequest) process(p *Participant) {
	// add hopefuls to any available slots
	// quorum is already locked by compile()
	i := 0
	for i < common.QuorumSize {
		if p.quorum.siblings[i] == nil {
			// perhaps there is a better way to do this???
			// also, there is an attack this opens up where someone submits you as a
			// new participant, and you accept yourself, and so your old self gets
			// thrown out =/
			//
			// that gets fixed once we start adding signatures to all updates
			if j.Sibling.compare(p.self) {
				p.self.index = byte(i)
			}

			j.Sibling.index = byte(i)
			p.heartbeats[j.Sibling.index] = make(map[siacrypto.TruncatedHash]*heartbeat)
			p.quorum.siblings[j.Sibling.index] = &j.Sibling

			println("placed hopeful at index", i)
			break
		}
		i++
	}
}

func (j *JoinRequest) GobEncode() (gobJR []byte, err error) {
	if j == nil {
		err = fmt.Errorf("Cannot encode a nil JoinRequest")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(j.Sibling)
	if err != nil {
		return
	}
	gobJR = w.Bytes()
	return
}

func (j *JoinRequest) GobDecode(gobJR []byte) (err error) {
	if j == nil {
		err = fmt.Errorf("Cannot decode into nil JoinRequest")
		return
	}

	r := bytes.NewBuffer(gobJR)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&j.Sibling)
	if err != nil {
		return
	}
	return
}
