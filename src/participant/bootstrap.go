package participant

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"network"
	"quorum"
	"quorum/script"
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

// currently a static variable, eventually there will be an entire process for
// finding address to bootstrap to.
var bootstrapAddress = network.Address{
	ID:   1,
	Host: "localhost",
	Port: 9988,
}

// CreateParticipant initializes a participant, and then either sets itself up
// as the bootstrap or establishes itself as a sibling on an existing network
func CreateParticipant(messageRouter network.MessageRouter, participantPrefix string) (p *Participant, err error) {
	// check for non-nil messageRouter
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// create a signature keypair for this participant
	pubKey, secKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// initialize State with default values and keypair
	p = &Participant{
		messageRouter: messageRouter,
		secretKey:     secKey,
		currentStep:   1,
	}

	// Can this be merged into one step?
	address := messageRouter.Address()
	address.ID = messageRouter.RegisterHandler(p)
	p.self = quorum.NewSibling(address, pubKey)

	// initialize heartbeat maps
	for i := range p.heartbeats {
		p.heartbeats[i] = make(map[siacrypto.TruncatedHash]*heartbeat)
	}

	// initialize disk variables
	p.quorum.SetWalletPrefix(participantPrefix)
	p.activeHistoryStep = SnapshotLen // trigger cylcing on the history during the first save

	// if we are the bootstrap participant, initialize a new quorum
	if p.self.Address() == bootstrapAddress {
		// create the bootstrap wallet
		err = p.quorum.CreateWallet(1, 4000, 0, 0, nil)
		if err != nil {
			return
		}

		p.quorum.AddSibling(p.self)
		p.newSignedHeartbeat()
		go p.tick()
		return
	}

	////////////////////////////
	// Bootstrap As A Hopeful //
	////////////////////////////

	// 1. Subscribe and start building blocks

	// 2. Download the quorum
	q := new(quorum.Quorum)
	fmt.Println("Getting Quorum Snapshot From Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.RecentSnapshot",
		Args: struct{}{},
		Resp: q,
	})
	if err != nil {
		return
	}
	p.quorum = *q

	// 3. Download the wallet list
	var walletList []quorum.WalletID
	fmt.Println("Getting List of Wallets From Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.SnapshotWalletList",
		Args: q.CurrentSnapshot(),
		Resp: &walletList,
	})

	// 4. Download the wallets
	var encodedWallets [][]byte
	fmt.Println("Getting all of the Wallets")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.SnapshotWallets",
		Args: SnapshotWalletsInput{
			Snapshot: q.CurrentSnapshot(),
			Ids:      walletList,
		},
		Resp: &encodedWallets,
	})
	if err != nil {
		return
	}

	println("ASGARD!")
	println(len(encodedWallets))

	for i, encodedWallet := range encodedWallets {
		println("JUST GO WITH IT")
		println(len(encodedWallet))
		err = p.quorum.LoadWallet(encodedWallet, walletList[i])
		if err != nil {
			return
		}
	}

	// 5. Download the blocks

	// 6. Fast forward the quorum from step 2

	// 7. Integrate with step 1 blocks, fast forward to immediate.

	// 7a Synchronize the participants - timing, step, currentSnapshot

	// 8. Submit a join quorum request

	// 9. The existing compile will handle the rest.

	// send a listener request to the bootstrap to become a listener on the quorum
	fmt.Println("Synchronizing to Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.Subscribe",
		Args: p.self.Address(),
		Resp: nil,
	})
	if err != nil {
		return
	}

	// Synchronize to the current quorum
	synchronize := new(Synchronize)
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.Synchronize",
		Args: struct{}{},
		Resp: synchronize,
	})
	if err != nil {
		return
	}

	// lock not needed as this is the only thread
	p.currentStep = synchronize.currentStep
	p.heartbeats = synchronize.heartbeats

	go p.tick()

	// encode an address and public key for script input
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	encoder.Encode(p.self.Address())
	encoder.Encode(pubKey)
	gobSibling := w.Bytes()

	// simple script that calls AddSibling
	var s script.ScriptInput
	s.Input = []byte{0x29, 0x04, byte(len(gobSibling)), 0xFF}
	s.Input = append(s.Input, gobSibling...)
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: s,
		Resp: nil,
	})

	return
}
