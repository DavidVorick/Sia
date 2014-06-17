package participant

import (
	"fmt"
	"network"
	"quorum"
	"quorum/script"
	"siacrypto"
	"siaencoding"
	"time"
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

		// create the bootstrap wallet
		p.quorum.CreateBootstrapWallet(BootstrapID, quorum.NewBalance(0, 15000), script.BootstrapScript)

		// create the add sibling script input
		var encSibling []byte
		encSibling, err = p.self.GobEncode()
		if err != nil {
			return
		}

		// execute the bootstrapping script
		si := script.ScriptInput{
			WalletID: BootstrapID,
			Input:    append(script.AddSiblingInput, encSibling...),
		}
		_, err = si.Execute(&p.quorum)
		if err != nil {
			return
		}

		siblings := p.quorum.Siblings()
		if siblings[0] == nil {
			err = fmt.Errorf("failed to add self to quorum")
			return
		}
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

	// 7. Request wallet from bootstrap
	walletID := siacrypto.RandomUInt64()
	encWalletID := siaencoding.EncUint64(walletID)
	in := append(encWalletID, script.DefaultScript...)
	s := script.ScriptInput{
		WalletID: BootstrapID,
		Input:    append(script.CreateWalletInput, in...),
	}

	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: s,
		Resp: nil,
	})
	if err != nil {
		return
	}

	// 8. Wait for next compile
	time.Sleep(time.Duration(quorum.QuorumSize) * StepDuration)

	// 9. Submit AddSibling request
	gobSibling, err := p.self.GobEncode()
	if err != nil {
		return
	}
	s = script.ScriptInput{
		WalletID: quorum.WalletID(walletID),
		Input:    append(script.AddSiblingInput, gobSibling...),
	}

	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: s,
		Resp: nil,
	})

	return
}
