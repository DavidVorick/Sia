package consensus

import (
	"errors"
	"sort"
	"time"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// An Update is the set of information sent by each participant during
// consensus. This information includes the heartbeat, which contains required
// information for being a part of the quorum, and it contains optional
// information such as script inputs.
type Update struct {
	Heartbeat          delta.Heartbeat
	HeartbeatSignature siacrypto.Signature

	ScriptInputs          []delta.ScriptInput
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
	errSignatoryMismatch = errors.New("signedUpdate has different number of signatures and signatories")
	errInvalidParent     = errors.New("signedUpdate targets a different block and/or quorum than the parent of this participant")
	errOutOfSync         = errors.New("update is late - not enough signatures given the current step of consensus")
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
	// Lock the engine and the updates variables
	p.engineLock.RLock()
	defer p.engineLock.RUnlock()

	p.updatesLock.Lock()
	defer p.updatesLock.Unlock()

	// Set the height and parent of the block.
	b.Height = p.engine.Metadata().Height
	b.ParentBlock = p.engine.Metadata().ParentBlock

	// Condense updates into a single non-repetitive block.
	{
		// Create a map containing all ScriptInputs found in a heartbeat.
		scriptInputMap := make(map[string]delta.ScriptInput)
		updateAdvancementMap := make(map[string]state.UpdateAdvancement)
		advancementSignatureMap := make(map[string]siacrypto.Signature)
		for i := range p.updates {
			if len(p.updates[i]) == 1 {
				for _, u := range p.updates[i] {
					// Add the heartbeat to the block
					b.Heartbeats[i] = u.Heartbeat
					b.HeartbeatSignatures[i] = u.HeartbeatSignature

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
						verified, err := p.engine.Metadata().Siblings[ua.Index].PublicKey.VerifyObject(u.AdvancementSignatures[i], ua)
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
	return
}

func (p *Participant) newSignedUpdate() {
	// Generate the entropy for this round of random numbers.
	var entropy state.Entropy
	copy(entropy[:], siacrypto.RandomByteSlice(state.EntropyVolume))

	hb := delta.Heartbeat{
		ParentBlock: p.engine.Metadata().ParentBlock,
		Entropy:     entropy,
		// storage proof
	}

	signature, err := p.secretKey.SignObject(hb)
	if err != nil {
		panic(err)
	}

	// Create the update with the heartbeat and heartbeat signature.
	update := Update{
		Heartbeat:          hb,
		HeartbeatSignature: signature,
	}

	// Attach all of the script inputs to the update, clearing the list of
	// script inputs in the process.
	p.scriptInputsLock.Lock()
	update.ScriptInputs = p.scriptInputs
	p.scriptInputs = nil
	p.scriptInputsLock.Unlock()

	// Attach all of the update advancements to the signed heartbeat and sign
	// them.
	p.updateAdvancementsLock.Lock()
	update.UpdateAdvancements = p.updateAdvancements
	p.updateAdvancements = nil
	for i, ua := range update.UpdateAdvancements {
		uas, err := p.secretKey.SignObject(ua)
		if err != nil {
			// log an error
			continue
		}
		update.AdvancementSignatures[i] = uas
	}
	p.updateAdvancementsLock.Unlock()

	// Sign the update and create a SignedUpdate object with ourselves as the
	// first signatory.
	updateSignature, err := p.secretKey.SignObject(update)
	su := SignedUpdate{
		Update:      update,
		Signatories: make([]byte, 1),
		Signatures:  make([]siacrypto.Signature, 1),
	}
	su.Signatories[0] = p.siblingIndex
	su.Signatures[0] = updateSignature

	// Add the heartbeat to our own heartbeat map.
	updateHash, err := siacrypto.HashObject(update)
	if err != nil {
		panic(err)
	}
	p.updates[p.siblingIndex][updateHash] = update

	// Broadcast the SignedUpdate to the network.
	p.broadcast(network.Message{
		Proc: "Participant.HandleSignedUpdate",
		Args: su,
	})
}

// TODO: add docstring
func (p *Participant) HandleSignedUpdate(su SignedUpdate, _ *struct{}) (err error) {
	// for debugging purposes
	defer func() {
		if err != nil && err != errHaveHeartbeat {
			println(err.Error())
		}
	}()

	// Lock the engine for the duration of the function.
	p.engineLock.RLock()
	defer p.engineLock.RUnlock()

	// Lock the updates variable for the duration of the function.
	p.updatesLock.Lock()
	defer p.updatesLock.Unlock()

	// Check that there is a signatory for every signature.
	if len(su.Signatories) != len(su.Signatures) {
		err = errSignatoryMismatch
		return
	}

	// Check that the Update matches the current block.
	// If it doesn't, it has one step to match the next block.
	if su.Update.Heartbeat.ParentBlock != p.engine.Metadata().ParentBlock {
		time.Sleep(StepDuration)
		if su.Update.Heartbeat.ParentBlock != p.engine.Metadata().ParentBlock {
			err = errInvalidParent
			return
		}
	}

	// Check that there are enough signatures in the update to match the
	// current step.
	p.currentStepLock.Lock()
	if int(p.currentStep) > len(su.Signatures) {
		p.currentStepLock.Unlock()
		err = errOutOfSync
		return
	}
	p.currentStepLock.Unlock()

	// Check that all of the signatures are valid, and that there are no repeats.
	updateHash, err := siacrypto.HashObject(su.Update)
	if err != nil {
		return
	}
	message := updateHash[:]
	previousSignatories := make(map[byte]bool)
	for i, signatory := range su.Signatories {
		// Check bounds on current signatory.
		if signatory >= state.QuorumSize {
			err = errBounds
			return
		}

		// Check that current signatory is a valid sibling in the quorum.
		if !p.engine.Metadata().Siblings[signatory].Active {
			err = errNonSibling
			return
		}

		// Check that current signatory has only been seen once in the current
		// SignedUpdate
		if previousSignatories[signatory] {
			err = errDoubleSign
			return
		}
		previousSignatories[signatory] = true

		// Verify the signature.
		verification := p.engine.Metadata().Siblings[signatory].PublicKey.Verify(su.Signatures[i], message)
		if !verification {
			err = errInvalidSignature
			return
		}

		// Extend the signed message so that it contians the proper message for
		// the next verification.
		message = append(su.Signatures[i][:], message...)
	}

	// Check if this update has already been received.
	_, exists := p.updates[su.Signatories[0]][updateHash]
	if exists {
		err = errHaveHeartbeat
		return
	}

	// Check that there are less than two heartbeats from this host yet seen.
	if len(p.updates[su.Signatories[0]]) >= 2 {
		err = errManyHeartbeats
		return
	}

	// Add the update to the list of seen updates.
	p.updates[su.Signatories[0]][updateHash] = su.Update

	// Sign the stack of signatures and append the signature to the stack, then
	// announce the Update to everyone on the quorum
	signature, err := p.secretKey.Sign(message)
	if err != nil {
		return
	}
	su.Signatures = append(su.Signatures, signature)
	su.Signatories = append(su.Signatories, p.siblingIndex)

	// broadcast the update to the quorum
	return
}
