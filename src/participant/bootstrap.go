package participant

import (
	//"bytes"
	//"encoding/gob"
	"fmt"
	"network"
	"quorum"
	"quorum/script"
	"siacrypto"
)

const (
	BootstrapID = 24
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
		p.heartbeats[i] = make(map[siacrypto.Hash]*heartbeat)
	}

	// initialize disk variables
	p.recentBlocks = make(map[uint32]*block)
	p.quorum.SetWalletPrefix(participantPrefix)
	p.activeHistoryStep = SnapshotLen // trigger cycling on the history during the first save

	// if we are the bootstrap participant, initialize a new quorum
	if p.self.Address() == bootstrapAddress {
		p.synchronized = true

		// the bootstrapping script
		// accepts two types of input:
		// - run script: 0 followed by key
		// - create wallet: non-zero followed by script body
		bScript := []byte{
			0x27, 0x01, //       load first byte of input
			0x1F, 0x00, 0x02, // if byte == 0
			0x2F, //                 move instruction pointer to input
			//                   else
			0x01, 0x01, //           push 0
			0x01, 0x64, //           push 100
			0x27, 0x08, //           push 8 bytes of input
			0x2D, 0x01, //           read rest of input into buffer 1
			0x30, 0x01, //           call create wallet
		}

		// create the bootstrap wallet
		p.quorum.CreateBootstrapWallet(BootstrapID, quorum.NewBalance(0, 15000), bScript)

		// create the bootstrapping script
		var encSibling []byte
		encSibling, err = p.self.GobEncode()
		if err != nil {
			return
		}

		slenh, slenl := byte(len(encSibling)<<8), byte(len(encSibling))
		sibScript := append([]byte{
			0x00,             //   zero byte indicates this is a run script request
			0x25, 0x00, 0x0B, //   move data pointer to start of encSibling
			0x02, slenh, slenl, // push encoding sibling length
			0x2B, 0x01, //         copy encoded sibling into buffer 1
			0x31, 0x01, //         call addSibling on buffer 1
			0xFF,
		}, encSibling...)

		// execute the bootstrapping script
		si := &script.ScriptInput{
			WalletID: BootstrapID,
			Input:    sibScript,
		}
		_, err = si.Execute(&p.quorum)
		if err != nil {
			return
		}

		siblings := p.quorum.Siblings()
		p.self = siblings[0]
		p.newSignedHeartbeat()
		go p.tick()
		return
	}

	////////////////////////////
	// Bootstrap As A Hopeful //
	////////////////////////////

	// 1. Synchronize to the current quorum to correctly produce blocks from
	// heartbeats
	synchronize := new(Synchronize)
	fmt.Println("Synchronizeing to the Bootstrap")
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

	// 2. Subscribe to the current quorum and receive all heartbeats
	fmt.Println("Subscribing to the Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.Subscribe",
		Args: p.self.Address(),
		Resp: nil,
	})
	if err != nil {
		return
	}

	// begin processing heartbeats
	go p.tick()

	// 3. Download a recent quorum snapshot
	fmt.Println("Getting Quorum Snapshot From Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.RecentSnapshot",
		Args: struct{}{},
		Resp: &p.quorum,
	})
	if err != nil {
		return
	}

	// 4. Download the wallet list
	var walletList []quorum.WalletID
	fmt.Println("Getting List of Wallets From Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.SnapshotWalletList",
		Args: p.quorum.CurrentSnapshot(),
		Resp: &walletList,
	})

	println("got wallet list")
	fmt.Println(walletList)

	// 5. Download the wallets
	var encodedWallets [][]byte
	fmt.Println("Getting all of the Wallets")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.SnapshotWallets",
		Args: SnapshotWalletsInput{
			Snapshot: p.quorum.CurrentSnapshot(),
			Ids:      walletList,
		},
		Resp: &encodedWallets,
	})
	if err != nil {
		return
	}

	for i, encodedWallet := range encodedWallets {
		err = p.quorum.InsertWallet(encodedWallet, walletList[i])
		if err != nil {
			return
		}
	}

	fmt.Println("Untouched Snapshot Status():")
	fmt.Println(p.quorum.Status())

	// 6. Download the blocks
	var blockList []block
	fmt.Println("Getting Blocks Since Snapshot")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.SnapshotBlocks",
		Args: p.quorum.CurrentSnapshot(),
		Resp: &blockList,
	})
	if err != nil {
		return
	}

	// 7. Integrate with blocks built while listening, compile all blocks
	for i := range blockList {
		p.appendBlock(&blockList[i])
	}

	currentHeight := p.quorum.Height()
	for p.recentBlocks[currentHeight] != nil {
		fmt.Println("Fast forwarding a block:")
		p.compile(p.recentBlocks[currentHeight])
		currentHeight += 1
	}
	p.synchronized = true // now compile will be called upon receiving a block

	/*
		// encode an address and public key for script input
		w := new(bytes.Buffer)
		encoder := gob.NewEncoder(w)
		encoder.Encode(p.self.Address())
		encoder.Encode(pubKey)
		gobSibling := w.Bytes()

		// simple script that calls AddSibling
		var s script.ScriptInput
		//s.WalletID = 1
		s.Input = []byte{0x29, 0x04, byte(len(gobSibling)), 0xFF}
		s.Input = append(s.Input, gobSibling...)
		err = p.messageRouter.SendMessage(&network.Message{
			Dest: bootstrapAddress,
			Proc: "Participant.AddScriptInput",
			Args: s,
			Resp: nil,
		})
	*/
	return
}
