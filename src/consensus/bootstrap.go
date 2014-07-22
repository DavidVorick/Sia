package consensus

import (
	"fmt"
	"network"
	"state"
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

// CreateBootstrapParticipant returns a participant that is participating as
// the first and only sibling on a new quorum.
func CreateBootstrapParticipant(mr network.MessageRouter, filePrefix string, sibID state.WalletID) (p *Participant, err error) {
	if sibID == 0 {
		err = fmt.Errorf("Cannot use id '0', this id is reserved for the bootstrapping wallet.")
		return
	}

	// Call NewParticipant, which gives a participant that has all of the basic fields initialized.
	p, err = NewParticipant(mr, filePrefix)
	if err != nil {
		return
	}

	// Call NewBootstrapEngine, which returns an engine that has a quorum with a
	// bootstrap/fountain wallet, and a sibling as described by p.self. The
	// sibling has also been given some funds.
	sib := state.Sibling{
		Address:   p.address,
		PublicKey: p.publicKey,
		WalletID:  sibID,
	}
	err = p.engine.Bootstrap(sib)
	if err != nil {
		return
	}

	// Set synchronized to true and start ticking.
	p.siblingIndex = 0
	p.synchronized = true
	go p.tick() // Tick gets its own thread, so the this function can return.

	return
}

func CreateJoiningParticipant(mr network.MessageRouter, filePrefix string, trustedSiblings []network.Address) (p *Participant, err error) {
	p, err = NewParticipant(mr, filePrefix)
	if err != nil {
		return
	}

	// An important step that is being omitted for this version of Sia (omitted
	// until Sia has meta-quorums and network-discovery) is verifying the hashes
	// of the snapshot and blocks before you actually attempt to acquire
	// anything.

	// 1. Submit a join request to the existing quorum. This join request will be
	// added to the heartbeats of the siblings, and will be included in the next
	// round of consensus. So before the join request gets through, the current
	// block will need to finish, and then the next block will need to finish as
	// well. This will take many hours, as block times are very slow.

	// 2. While waiting for the next block, can download a snapshot and all
	// blocks following the snapshot. The 3 items of concern are: Metadata,
	// Wallets, Events.

	// 3. After bringing the quorum up to date (still missing the latest block,
	// won't be able to self-compile), can begin downloading file segments. The
	// only wallet segements to avoid are the wallet segments with active
	// uploads. Since you aren't announced to the quorum yet, the uploader won't
	// know to contact you and upload to you the file diff.

	// 4. After being accepted to the quorum as a full sibling, all downloads are
	// fair game.

	// 5. After collecting all downloads, announce synchronization and switch
	// from being an unpaid bootstrapping participant to a paid active
	// participant.

	return
}

/* // CreateParticipant initializes a participant, and then either sets itself up
// as the bootstrap or establishes itself as a sibling on an existing network
func CreateParticipant(messageRouter network.MessageRouter, participantPrefix string, bootstrap bool) (p *Participant, err error) {

	////////////////////////////
	// Bootstrap As A Hopeful //
	////////////////////////////

	// 1. Synchronize to the current quorum to correctly produce blocks from
	// heartbeats
	synchronize := new(Synchronize)
	fmt.Println("Synchronizing to the Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: BootstrapAddress,
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
		Dest: BootstrapAddress,
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
		Dest: BootstrapAddress,
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
		Dest: BootstrapAddress,
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
		Dest: BootstrapAddress,
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
	var blockList []delta.Block
	fmt.Println("Getting Blocks Since Snapshot")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: BootstrapAddress,
		Proc: "Participant.SnapshotBlocks",
		Args: p.quorum.CurrentSnapshot(),
		Resp: &blockList,
	})
	if err != nil {
		return
	}

	// 7. Integrate with blocks built while listening, compile all blocks
	//for i := range blockList {
	// p.appendBlock(&blockList[i])
	//}

	currentHeight := p.quorum.Height()
	for p.recentBlocks[currentHeight] != nil {
		fmt.Println("Fast forwarding a block:")
		p.compile(p.recentBlocks[currentHeight])
		currentHeight += 1
	}
	p.synchronized = true // now compile will be called upon receiving a block

	// 8. Request wallet from bootstrap
	walletID := siacrypto.RandomUint64()
	s := script.ScriptInput{
		WalletID: BootstrapID,
		Input:    script.CreateWalletInput(walletID, script.DefaultScript(p.self.PublicKey())),
	}

	err = p.messageRouter.SendMessage(&network.Message{
		Dest: BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: s,
		Resp: nil,
	})
	if err != nil {
		return
	}

	// 9. Wait for next compile
	time.Sleep(time.Duration(quorum.QuorumSize) * StepDuration)

	// 10. Create and send AddSibling request
	gobSibling, err := p.self.GobEncode()
	if err != nil {
		return
	}
	input, err := script.SignInput(p.secretKey, script.AddSiblingInput(gobSibling))
	if err != nil {
		return
	}
	s = script.ScriptInput{
		WalletID: quorum.WalletID(walletID),
		Input:    input,
	}

	err = p.messageRouter.SendMessage(&network.Message{
		Dest: BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: s,
		Resp: nil,
	})

	return
}*/
