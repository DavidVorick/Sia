package consensus

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

const (
	NumSteps = state.QuorumSize + 1
)

// An Update is the set of information sent by each participant during
// consensus. This information includes the heartbeat, which contains required
// information for being a part of the quorum, and it contains optional
// information such as script inputs.
type Update struct {
	Height             uint32
	Heartbeat          delta.Heartbeat
	HeartbeatSignature siacrypto.Signature

	ScriptInputs          []state.ScriptInput
	UpdateAdvancements    []state.UpdateAdvancement
	AdvancementSignatures []siacrypto.Signature
}

// TODO: add docstring
type SignedUpdate struct {
	Update      Update // eventually be replaced with a hash and fetch request
	Signatories []byte
	Signatures  []siacrypto.Signature
}

var (
	errNotReady          = errors.New("not ready to receive heartbeats yet")
	errSignatoryMismatch = errors.New("signedUpdate has different number of signatures and signatories")
	errInvalidParent     = errors.New("signedUpdate targets a different block and/or quorum than the parent of this participant")
	errLateUpdate        = errors.New("update is late - not enough signatures given the current step of consensus")
	errBounds            = errors.New("update contains a signature from an out-of-bounds signatory")
	errNonSibling        = errors.New("update contains a signature from a non-sibling")
	errDoubleSign        = errors.New("update contains two signatures from the same signatory")
	errInvalidSignature  = errors.New("update contains a corrupted/invalid signature")
	errHaveHeartbeat     = errors.New("update has already been processed")
	errManyHeartbeats    = errors.New("multiple heartbeats from this sibling have already been submitted")
)

// condenseBlock assumes that a heartbeat has a valid signature and that the
// parent is the correct parent.
func (p *Participant) condenseBlock() (b delta.Block) {
	// Set the height and parent of the block.
	p.engineLock.RLock()
	b.Height = p.engine.Metadata().Height
	b.ParentBlock = p.engine.Metadata().ParentBlock
	p.engineLock.RUnlock()

	// Condense updates into a single non-repetitive block.
	p.updatesLock.Lock()
	{
		// Create a map containing all ScriptInputs found in a heartbeat.
		scriptInputMap := make(map[string]state.ScriptInput)
		updateAdvancementMap := make(map[string]state.UpdateAdvancement)
		advancementSignatureMap := make(map[string]siacrypto.Signature)
		for i := range p.updates {
			if len(p.updates[i]) == 1 {
				for _, u := range p.updates[i] {
					// Add the heartbeat to the block for active siblings.
					p.engineLock.RLock()
					if p.engine.Metadata().Siblings[i].Active() {
						b.Heartbeats[i] = u.Heartbeat
						b.HeartbeatSignatures[i] = u.HeartbeatSignature
					}
					p.engineLock.RUnlock()

					// Add all of the script inputs to the script input map.
					for _, scriptInput := range u.ScriptInputs {
						scriptInputHash, err := siacrypto.HashObject(scriptInput)
						if err != nil {
							continue
						}
						scriptInputMap[string(scriptInputHash[:])] = scriptInput
					}

					// Add all of the update advancements to the hash map.
					for i, ua := range u.UpdateAdvancements {
						// Verify the signature on the update advancement.
						verified, err := p.engine.Metadata().Siblings[ua.SiblingIndex].PublicKey.VerifyObject(u.AdvancementSignatures[i], ua)
						if err != nil || !verified {
							continue
						}
						uaHash, err := siacrypto.HashObject(ua)
						if err != nil {
							continue
						}
						uaString := string(uaHash[:])
						updateAdvancementMap[uaString] = ua
						advancementSignatureMap[uaString] = u.AdvancementSignatures[i]
					}
				}
			}

			// Clear the update map for this sibling, so that it is clean during the
			// next round of consensus.
			p.updates[i] = make(map[siacrypto.Hash]Update)
		}

		// Sort the scriptInputMap and include the scriptInputs into the block in
		// sorted order.
		var sortedKeys []string
		for k := range scriptInputMap {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		for _, k := range sortedKeys {
			b.ScriptInputs = append(b.ScriptInputs, scriptInputMap[k])
		}

		// Sort the updateAdvancementMap and include the advancements into the
		// block in sorted order.
		sortedKeys = nil
		for k := range updateAdvancementMap {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		for _, k := range sortedKeys {
			b.UpdateAdvancements = append(b.UpdateAdvancements, updateAdvancementMap[k])
			b.AdvancementSignatures = append(b.AdvancementSignatures, advancementSignatureMap[k])
		}
	}
	p.updatesLock.Unlock()
	return
}

// newSignedUpdate creates an update for this participant, signs it, and then
// broadcasts it to the network.
func (p *Participant) newSignedUpdate() {
	// Check that this function was not called by error.
	if p.engine.SiblingIndex() > state.QuorumSize {
		p.log.Error("error call on newSignedUpdate")
		return
	}

	// Generate the entropy for this round of random numbers.
	var entropy state.Entropy
	copy(entropy[:], siacrypto.RandomByteSlice(state.EntropyVolume))

	sp, err := p.engine.BuildStorageProof()
	if err == state.ErrEmptyQuorum {
		p.log.Debug("could not build storage proof:", err)
	} else if err != nil {
		p.log.Error("failed to construct storage proof:", err)
		return
	}
	hb := delta.Heartbeat{
		ParentBlock:  p.engine.Metadata().ParentBlock,
		Entropy:      entropy,
		StorageProof: sp,
	}

	signature, err := p.secretKey.SignObject(hb)
	if err != nil {
		p.log.Error("failed to sign heartbeat:", err)
		return
	}

	// Create the update with the heartbeat and heartbeat signature.
	p.engineLock.RLock()
	update := Update{
		Height:             p.engine.Metadata().Height,
		Heartbeat:          hb,
		HeartbeatSignature: signature,
	}
	p.engineLock.RUnlock()

	// Attach all of the script inputs to the update, clearing the list of
	// script inputs in the process.
	p.updatesLock.Lock()
	update.ScriptInputs = p.scriptInputs
	p.scriptInputs = nil
	p.updatesLock.Unlock()

	// Attach all of the update advancements to the signed heartbeat and sign
	// them.
	p.updatesLock.Lock()
	update.UpdateAdvancements = p.updateAdvancements
	p.updateAdvancements = nil
	for _, ua := range update.UpdateAdvancements {
		uas, err := p.secretKey.SignObject(ua)
		if err != nil {
			// log an error
			continue
		}
		update.AdvancementSignatures = append(update.AdvancementSignatures, uas)
	}
	p.updatesLock.Unlock()

	// Sign the update and create a SignedUpdate object with ourselves as the
	// first signatory.
	updateSignature, err := p.secretKey.SignObject(update)
	su := SignedUpdate{
		Update:      update,
		Signatories: make([]byte, 1),
		Signatures:  make([]siacrypto.Signature, 1),
	}
	su.Signatories[0] = p.engine.SiblingIndex()
	su.Signatures[0] = updateSignature

	// Add the heartbeat to our own heartbeat map.
	updateHash, err := siacrypto.HashObject(update)
	if err != nil {
		p.log.Error("failed to hash update:", err)
		return
	}
	p.updatesLock.Lock()
	p.updates[p.engine.SiblingIndex()][updateHash] = update
	p.updatesLock.Unlock()

	// Broadcast the SignedUpdate to the network.
	p.broadcast(network.Message{
		Proc: "Participant.HandleSignedUpdate",
		Args: su,
	})
}

// HandleSignedUpdate is an RPC that allows other hosts to submit updates with
// signatures to this host. They will be processed according to the rules of
// concensus, blocking late updates and waiting on early updates, and throwing
// out anything that does not follow the rules for legal signatures.
func (p *Participant) HandleSignedUpdate(su SignedUpdate, _ *struct{}) (err error) {
	// Debugging: Print the participant, their address, and which update they've been sent.

	// If ticking hasn't started yet, wait until tick() is called.
	p.tickLock.RLock()
	if !p.ticking {
		p.tickLock.RUnlock()

		// updateStop will be write-locked until the participant starts
		// ticking, meaning this function will block until p.ticking ==
		// true
		p.updateStop.RLock()
		p.updateStop.RUnlock()

		p.tickLock.RLock()
	}
	p.tickLock.RUnlock()

	// Printing errors helps with debugging. Production code for this
	// package should never print, only log.
	defer func() {
		if err != nil && err != errHaveHeartbeat {
			fmt.Println(err.Error())
		}
	}()

	// Check that there is a signatory for every signature.
	if len(su.Signatories) != len(su.Signatures) {
		err = errSignatoryMismatch
		return
	}

	// Check that the update is not late.
	p.tickLock.RLock()
	p.engineLock.RLock()
	if (su.Update.Height == p.engine.Metadata().Height && int(p.currentStep) > len(su.Signatures)) || su.Update.Height < p.engine.Metadata().Height {
		err = errLateUpdate
		p.tickLock.RUnlock()
		p.engineLock.RUnlock()
		return
	}
	p.tickLock.RUnlock()
	p.engineLock.RUnlock()

	// If the update is early, wait until the proper block appears. We
	// don't want to process an update earlier than step 1 because other
	// siblings in the network (due to clock drift) may not be far enough
	// along to handle it, and we want to give the update time to
	// propagate.
	p.tickLock.RLock()
	p.engineLock.RLock()
	for su.Update.Height > p.engine.Metadata().Height || (su.Update.Height == p.engine.Metadata().Height && p.currentStep < 1) {
		// Sleep until the next step, repeating until the height has properly caught up.
		timeRemainingThisStep := StepDuration - (time.Since(p.tickStart) % StepDuration)

		// Unlock all mutexes, sleep, and then relock all mutexes.
		p.engineLock.RUnlock()
		p.tickLock.RUnlock()
		time.Sleep(timeRemainingThisStep + 5*time.Millisecond) // 5 extra milliseconds for good luck.
		p.engineLock.RLock()
		p.tickLock.RLock()

	}
	p.tickLock.RUnlock()
	p.engineLock.RUnlock()

	// Check that all of the signatures are valid, and that there are no repeats.
	p.engineLock.RLock()
	p.updatesLock.Lock()
	updateHash, err := siacrypto.HashObject(su.Update)
	if err != nil {
		p.updatesLock.Unlock()
		p.engineLock.RUnlock()
		return
	}
	message := updateHash[:]
	previousSignatories := make(map[byte]bool)
	for i, signatory := range su.Signatories {
		// Check bounds on current signatory.
		if signatory >= state.QuorumSize {
			err = errBounds
			p.updatesLock.Unlock()
			p.engineLock.RUnlock()
			return
		}

		// Check that current signatory is a valid sibling in the quorum.
		if p.engine.Metadata().Siblings[signatory].Inactive() {
			err = errNonSibling
			p.updatesLock.Unlock()
			p.engineLock.RUnlock()
			return
		}

		// Check that current signatory has only been seen once in the current
		// SignedUpdate
		if previousSignatories[signatory] {
			err = errDoubleSign
			p.updatesLock.Unlock()
			p.engineLock.RUnlock()
			return
		}
		previousSignatories[signatory] = true

		// Verify the signature.
		verification := p.engine.Metadata().Siblings[signatory].PublicKey.Verify(su.Signatures[i], message)
		if !verification {
			err = errInvalidSignature
			p.updatesLock.Unlock()
			p.engineLock.RUnlock()
			return
		}

		// Extend the signed message so that it contians the proper message for
		// the next verification.
		message = append(su.Signatures[i][:], message...)
	}
	p.engineLock.RUnlock()

	// Check if this update has already been received.
	_, exists := p.updates[su.Signatories[0]][updateHash]
	if exists {
		err = errHaveHeartbeat
		p.updatesLock.Unlock()
		return
	}

	// Check that there are less than two heartbeats from this host yet seen.
	if len(p.updates[su.Signatories[0]]) >= 2 {
		err = errManyHeartbeats
		p.updatesLock.Unlock()
		return
	}

	// Add the update to the list of seen updates.
	p.updates[su.Signatories[0]][updateHash] = su.Update
	p.updatesLock.Unlock()

	// Sign the stack of signatures and append the signature to the stack, then
	// announce the Update to everyone on the quorum
	signature, err := p.secretKey.Sign(message)
	if err != nil {
		return
	}
	p.engineLock.RLock()
	su.Signatures = append(su.Signatures, signature)
	su.Signatories = append(su.Signatories, p.engine.SiblingIndex())
	p.engineLock.RUnlock()

	// broadcast the update to the quorum
	return
}
