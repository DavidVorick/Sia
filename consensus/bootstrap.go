package consensus

import (
	"errors"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
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
		err = errors.New("cannot use id '0', this id is reserved for the bootstrapping wallet")
		return
	}

	// create basic participant
	p, err = NewParticipant(mr, filePrefix)
	if err != nil {
		return
	}
	p.siblingIndex = 0

	// create a bootstrap wallet, and a wallet for this participant to use
	err = p.engine.Bootstrap(state.Sibling{
		Address:   p.address,
		PublicKey: p.publicKey,
		WalletID:  sibID,
	})
	if err != nil {
		return
	}

	// set synchronized to true and start ticking
	p.synchronized = true
	go p.tick()

	return
}

// TODO: add docstring
func CreateJoiningParticipant(mr network.MessageRouter, filePrefix string, tetherID state.WalletID, quorumSiblings []network.Address) (p *Participant, err error) {
	p, err = NewParticipant(mr, filePrefix)
	if err != nil {
		return
	}

	// An important step that is being omitted for this version of Sia (omitted
	// until Sia has meta-quorums and network-discovery) is verifying the
	// hashes of the snapshot and blocks before you actually attempt to acquire
	// anything.

	// There is an assumption that the input wallet exists on the quorum with a
	// balance sufficient to cover the costs of creating the participant.

	// 1. Submit a join request to the existing quorum. This join request will
	// be added to the heartbeats of the siblings, and will be included in the
	// next round of consensus. So before the join request gets through, the
	// current block will need to finish, and then the next block will need to
	// finish as well. This will take many hours, as block times are very slow.
	joinRequest := delta.ScriptInput{
		WalletID: tetherID,
		Input:    delta.DefaultScript(p.publicKey),
	}
	for _, address := range quorumSiblings {
		mr.SendAsyncMessage(network.Message{
			Dest: address,
			Proc: "Participant.AddScriptInput",
			Args: joinRequest,
		})
	}

	// 2. While waiting for the next block, download a snapshot.
	// The 3 items of concern are: Metadata, Wallets, Events.
	{
		// get height of the most recent snapshot
		var snapshotHead uint32
		mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.RecentSnapshotHeight",
			Args: struct{}{},
			Resp: &snapshotHead,
		})

		// get the metadata from the snapshot
		var snapshotMetadata state.Metadata
		mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.SnapshotMetadata",
			Args: snapshotHead,
			Resp: &snapshotMetadata,
		})
		p.engine.BootstrapSetMetadata(snapshotMetadata)

		// get the list of wallets in the snapshot
		var walletList []state.WalletID
		mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.SnapshotWalletList",
			Args: snapshotHead,
			Resp: &walletList,
		})

		// get each wallet individually and insert them into the quorum
		for _, walletID := range walletList {
			swa := SnapshotWalletArg{
				SnapshotHead: snapshotHead,
				WalletID:     walletID,
			}

			var wallet state.Wallet
			mr.SendMessage(network.Message{
				Dest: quorumSiblings[0],
				Proc: "Participant.SnapshotWallet",
				Args: swa,
				Resp: &wallet,
			})

			err = p.engine.BootstrapInsertWallet(wallet)
			if err != nil {
				// ???, panic would be inappropriate
			}
		}
	}

	// Events will be implemented at a later time.

	// 3. Download all of the blocks that have been processed since the
	// snapshot, which will bring the quorum up to date, except for being
	// behind in the current round of consensus.
	{
		// figure out which block height is the latest
		var currentMetadata state.Metadata
		mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &currentMetadata,
		})

		for p.engine.Metadata().Height < currentMetadata.Height-1 {
			var b delta.Block
			mr.SendMessage(network.Message{
				Dest: quorumSiblings[0],
				Proc: "Participant.Block",
				Args: p.engine.Metadata().Height,
				Resp: &b,
			})
			p.engine.Compile(b)
		}
	}

	// 4. After bringing the quorum up to date (still missing the latest block,
	// won't be able to self-compile), can begin downloading file segments. The
	// only wallet segements to avoid are the wallet segments with active
	// uploads. Since you aren't announced to the quorum yet, the uploader
	// won't know to contact you and upload to you the file diff.

	// 5. After being accepted to the quorum as a full sibling, all downloads
	// are fair game.

	// 6. After collecting all downloads, announce synchronization and switch
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
}*/
