package consensus

import (
	"errors"
	"time"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
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

// A helper function for CreateJoiningParticipant, that downloads the next
// block and compiles it into the participant. fetchAndCompileNextBlock
// requires the engine mutex to be locked.
func (p *Participant) fetchAndCompileNextBlock(quorumSiblings []network.Address) (err error) {
	var b delta.Block
	err = p.messageRouter.SendMessage(network.Message{
		Dest: quorumSiblings[0],
		Proc: "Participant.Block",
		Args: p.engine.Metadata().Height,
		Resp: &b,
	})
	if err != nil {
		return
	}

	p.engine.Compile(b)
	return
}

// CreateBootstrapParticipant returns a participant that is participating as
// the first and only sibling on a new quorum.
func CreateBootstrapParticipant(mr network.MessageRouter, filePrefix string, bootstrapTetherWallet state.WalletID, tetherWalletPublicKey siacrypto.PublicKey) (p *Participant, err error) {
	// ID 0 is reserved for the early-distribution 'fountain' wallet. The
	// full netowrk is not likely to have this, but it makes test-network
	// actions a lot simpler.
	if bootstrapTetherWallet == 0 {
		err = errors.New("cannot use id '0', this id is reserved for the bootstrapping wallet")
		return
	}

	// Create basic participant.
	p, err = newParticipant(mr, filePrefix)
	if err != nil {
		return
	}

	// Create a bootstrap sibling, using the bootstrapTether id as the
	// wallet id that the sibling will be tethered to.
	bootstrapSibling := state.Sibling{
		Address:   p.address,
		PublicKey: p.publicKey,
		WalletID:  bootstrapTetherWallet,
	}
	err = p.engine.Bootstrap(bootstrapSibling, tetherWalletPublicKey)
	if err != nil {
		return
	}
	p.siblingIndex = 0

	// Create the first update.
	p.newSignedUpdate()

	// Run the first compile.
	block := p.condenseBlock()
	p.engine.Compile(block)
	p.newSignedUpdate()

	// Begin ticking.
	go p.tick()

	return
}

// CreateJoiningParticipant creates a new participant and integrates it as a
// host with an existing quorum. It is assumed that the tetherID is an ID to a
// generic wallet, and that the secret key is the key that should be the key
// that is assiciated with the public key of the generic wallet.
func CreateJoiningParticipant(mr network.MessageRouter, filePrefix string, tetherID state.WalletID, tetherWalletSecretKey siacrypto.SecretKey, quorumSiblings []network.Address) (p *Participant, err error) {
	// Create a new, basic participant.
	p, err = newParticipant(mr, filePrefix)
	if err != nil {
		return
	}

	// An important step that is being omitted for this version of Sia
	// (omitted until Sia has meta-quorums and network-discovery) is
	// verifying the hashes of the snapshot and blocks before you actually
	// attempt to acquire anything.

	// There is an assumption that the input wallet exists on the quorum
	// with a balance sufficient to cover the costs of creating the
	// participant. This needs to be tested and verified.

	// Download a snapshot, which will form the basis of synchronization.
	{
		// get height of the most recent snapshot
		var metadata state.Metadata
		err = mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &metadata,
		})
		if err != nil {
			return
		}

		// get the metadata from the snapshot
		var snapshotMetadata state.Metadata
		err = mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.SnapshotMetadata",
			Args: metadata.RecentSnapshot,
			Resp: &snapshotMetadata,
		})
		if err != nil {
			return
		}
		p.engine.BootstrapSetMetadata(snapshotMetadata)

		// get the list of wallets in the snapshot
		var walletList []state.WalletID
		err = mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.SnapshotWalletList",
			Args: metadata.RecentSnapshot,
			Resp: &walletList,
		})
		if err != nil {
			return
		}

		// get each wallet individually and insert them into the quorum
		for _, walletID := range walletList {
			swa := SnapshotWalletArg{
				SnapshotHead: metadata.RecentSnapshot,
				WalletID:     walletID,
			}

			var wallet state.Wallet
			err = mr.SendMessage(network.Message{
				Dest: quorumSiblings[0],
				Proc: "Participant.SnapshotWallet",
				Args: swa,
				Resp: &wallet,
			})
			if err != nil {
				return
			}

			err = p.engine.BootstrapInsertWallet(wallet)
			if err != nil {
				// ???, panic would be inappropriate
			}
		}

		// Event downloading will be implemented later.
	}

	// At this point, saveBlock() in package delta is expecting the active
	// history file to be available, but this file hasn't been created yet
	// because the inital values for snapshot weren't established
	// correctly. It's a bit of a hack and should be refactored at some
	// point, but we've got to set those variables so that compile(),
	// saveBlock(), and saveSnapshot() work as expected.
	err = p.engine.BootstrapJoinSetup()
	if err != nil {
		return
	}

	// Download all of the blocks that have been processed since the
	// snapshot, which will bring the quorum up to date, except for being
	// behind in the current round of consensus.
	{
		// figure out which block height is the latest
		var currentMetadata state.Metadata
		err = mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &currentMetadata,
		})
		if err != nil {
			return
		}

		p.engineLock.Lock()
		for p.engine.Metadata().Height < currentMetadata.Height {
			err = p.fetchAndCompileNextBlock(quorumSiblings)
			if err != nil {
				return
			}
		}
		p.engineLock.Unlock()
	}

	// Synchronize to the quorum (this implementation is non-cryptographic)
	// and begin ticking.
	//
	// The method is a bit haphazard, but the goal is to insure that when
	// ticking starts and updates come in, the quorum is guaranteed to be
	// caught up. We do make certain (reasonable) assumptions about network
	// speed.
	//
	// 1. Get the synchronization information of the current network,
	// including the current height, the current step, and the progress
	// throug the current step.
	//
	// 2. Submit the join request, waiting if the step is 0-2 because it's
	// unclear whether the join request will make it into the current block
	// with the low step numbers.
	//
	// 3. Download any blocks that are missing, catching up to the current
	// quorum.
	//
	// 4. Wait for this block to finish (will not have join request), and
	// the next block to finish (will have join request).
	//
	// 5. Begin ticking as soon as the current quorum hits step 0 after the
	// block containint the join request. The participant can now start
	// receiving SignedUpdates without synchronization issues.
	//
	// 6. Download the two blocks that we're missing. Updates will be held
	// by the HandleSignedUpdate function until step 2 is reached, so these
	// two blocks must be downloaded before step 2 is reached. The most
	// recently completed block is not guaranteed to be available until
	// step 1 is reached, so this must be taken into consideration.
	//
	// That's it! The rest of the code should maintain synchronization.
	{
		var cps ConsensusProgressStruct
		err = mr.SendMessage(network.Message{
			Dest: quorumSiblings[0],
			Proc: "Participant.ConsensusProgress",
			Args: struct{}{},
			Resp: &cps,
		})
		if err != nil {
			return
		}
		cpsReceived := time.Now()

		// Don't submit the joinRequest unless the step is greater than
		// 1. It creates uncertainty over which block the join request
		// will be accepted in. This can be revisited later to remove
		// this artificial constraint.
		if cps.CurrentStep < 3 {
			time.Sleep(StepDuration * time.Duration(3-cps.CurrentStep))
		}

		// REMEMBER TO SET DEADLINE TO CPS.HEIGHT + 2!
		// Create the join request and send it to the quorum.
		var joinRequest delta.ScriptInput
		inputSibling := state.Sibling{
			Address:   p.address,
			PublicKey: p.publicKey,
		}
		joinRequest, err = delta.AddSiblingInput(tetherID, inputSibling, tetherWalletSecretKey)
		if err != nil {
			return
		}
		for _, address := range quorumSiblings {
			// Something should asynchronously log any errors
			// returned.
			mr.SendAsyncMessage(network.Message{
				Dest: address,
				Proc: "Participant.AddScriptInput",
				Args: joinRequest,
			})
		}

		// Download any blocks that are missing that are currently
		// available.
		for p.engine.Metadata().Height < cps.Height {
			p.engineLock.Lock()
			err = p.fetchAndCompileNextBlock(quorumSiblings)
			p.engineLock.Unlock()
			if err != nil {
				return
			}
		}

		// Wait for the current block to finish, and then for the next
		// block to also finish, and begin ticking when the following
		// block hits step 0.
		sleepDuration := (time.Duration(NumSteps-cps.CurrentStep) * StepDuration) - time.Since(cpsReceived) + time.Duration(NumSteps)*StepDuration - cps.CurrentStepProgress
		time.Sleep(sleepDuration)
		go p.tick()

		// Download the first missing block.
		p.engineLock.Lock()
		err = p.fetchAndCompileNextBlock(quorumSiblings)
		p.engineLock.Unlock()
		if err != nil {
			return
		}

		// Sleep another step so that the second block becomes
		// available.
		time.Sleep(StepDuration)

		// Download second missing block.
		p.engineLock.Lock()
		err = p.fetchAndCompileNextBlock(quorumSiblings)
		p.engineLock.Unlock()
		if err != nil {
			return
		}
	}

	// Parse the metadata and figure out which sibling is ourselves.
	p.engineLock.Lock()
	for i, sibling := range p.engine.Metadata().Siblings {
		if sibling.Address == p.address && sibling.PublicKey == p.publicKey {
			p.siblingIndex = byte(i)
			p.setSiblingIndex(p.siblingIndex)
			break
		}
	}
	p.engineLock.Unlock()

	// Once accepted as a sibling, begin downloading all files.
	// Be careful with overwrites regarding uploads that come to fruition. I think this is as simple as rejecting/ignoring updates until downloading is complete for the given file.
	// Once all files are downloaded, announce full siblingness.
	return
}
