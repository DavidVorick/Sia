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

	// Create basic participant.
	p, err = newParticipant(mr, filePrefix)
	if err != nil {
		return
	}
	p.siblingIndex = 0

	// Create a bootstrap wallet, and a wallet for this participant to use.
	err = p.engine.Bootstrap(state.Sibling{
		Address:   p.address,
		PublicKey: p.publicKey,
		WalletID:  sibID,
	})
	if err != nil {
		return
	}

	// Create the first update.
	p.newSignedUpdate()

	// Begin ticking.
	go p.tick()

	return
}

// TODO: add docstring
func CreateJoiningParticipant(mr network.MessageRouter, filePrefix string, tetherID state.WalletID, quorumSiblings []network.Address) (p *Participant, err error) {
	// Create a new, basic participant.
	p, err = newParticipant(mr, filePrefix)
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
	{
		// Create the sibling object that will be submitted to the
		// quorum.
		//inputSibling := state.Sibling{
		//	Address:   p.address,
		//	PublicKey: p.publicKey,
		//}

		// Create the join request and send it to the quorum.
		joinRequest := delta.ScriptInput{
			WalletID: tetherID,
			Input:    delta.AddSiblingInput(nil), //(inputSibling),
		}
		for _, address := range quorumSiblings {
			mr.SendAsyncMessage(network.Message{
				Dest: address,
				Proc: "Participant.AddScriptInput",
				Args: joinRequest,
			})
		}
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

		// Event downloading will be implemented later.
	}

	// Synchronize to the quorum (this implementation is non-cryptographic)
	// and being ticking. The first block will be unsuccessful, but this
	// gap will be filled by the block downloading.
	{
		var cps ConsensusProgressStruct
		mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.ConsensusProgress",
			Args: struct{}{},
			Resp: &cps,
		})

		// We would want to start ticking, but in a way that makes sure
		// we don't compile a block that we weren't around to properly
		// listen to.
		//
		// And really, we don't want to start ticking until we know
		// we've made it into the quorum.  Which potentially means that
		// we're going to be waiting extra blocks... >.<
		p.currentStep = cps.CurrentStep
	}

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

	// Before spending any bandwidth on downloading, make sure that we have
	// been accepted into the quorum. (hopefully) The expiration on the add
	// sibling message is set to only one block in the future, meaning if
	// we wait a full block cycle and aren't in the quorum, that we will
	// never be in the quorum. Return an error.

	// 4. After bringing the quorum up to date (still missing the latest block,
	// won't be able to self-compile), can begin downloading file segments. The
	// only wallet segements to avoid are the wallet segments with active
	// uploads. Since you aren't announced to the quorum yet, the uploader
	// won't know to contact you and upload to you the file diff. Do this
	// in a gothread.

	// 5. After being accepted to the quorum as a sibling, all downloads
	// are fair game. Complete these in a gothread.

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
