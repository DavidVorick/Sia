package consensus

import (
	"delta"
	"errors"
	"fmt"
	"network"
	"siacrypto"
	"state"
	"time"
)

type SignedUpdate struct {
	Update      Update // eventually be replaced with a hash and fetch request
	Signatories []byte
	Signatures  []siacrypto.Signature
}

var hsuerrSignatureSignatoryMismatch = errors.New("SignedUpdate has different number of signatures and signatories.")
var hsuerrInvalidParent = errors.New("SignedUpdate targets a different block and/or quorum than the parent of this participant.")
var hsuerrOutOfSync = errors.New("Update is late - not enough signatures given the current step of consensus.")
var hsuerrBounds = errors.New("Update contains a signature from an out-of-bounds signatory.")
var hsuerrNonSibling = errors.New("Update contains a signature from a non-sibling.")
var hsuerrDoubleSign = errors.New("Update contains two signatures from the same signatory.")
var hsuerrInvalidSignature = errors.New("Update contains a corrupted/invalid signature.")
var hsuerrHaveHeartbeat = errors.New("This update has already been processed.")
var hsuerrManyHeartbeats = errors.New("Multiple heartbeats from this sibling have already been submitted.")

func (p *Participant) NewSignedHeartbeat() {
	// Generate the entropy for this round of random numbers.
	var entropy state.Entropy
	randomBytes := siacrypto.RandomByteSlice(state.EntropyVolume)
	copy(entropy[:], randomBytes)

	hb := delta.Heartbeat{
		ParentBlock: p.engine.Metadata().ParentBlock,
		Entropy:     entropy,
		// storage proof
	}

	signature, err := p.secretKey.SignObject(hb)
	if err != nil {
		panic(err)
	}

	update := Update{
		Heartbeat:          hb,
		HeartbeatSignature: signature,
	}
	updateSignature, err := p.secretKey.SignObject(update)

	su := SignedUpdate{
		Update:      update,
		Signatories: make([]byte, 1),
		Signatures:  make([]siacrypto.Signature, 1),
	}
	su.Signatories[0] = p.siblingIndex
	su.Signatures[0] = updateSignature

	p.broadcast(network.Message{
		Proc: "Participant.HandleSignedHeartbeat",
		Args: su,
		Resp: err,
	})
}

// The series of printlns in this function are purely for debugging.
func (p *Participant) HandleSignedUpdate(su SignedUpdate, _ *struct{}) (err error) {
	// Lock the engine for the duration of the function.
	p.engineLock.RLock()
	defer p.engineLock.RUnlock()

	// Lock the updates variable for the duration of the function.
	p.updatesLock.Lock()
	defer p.updatesLock.Unlock()

	// Check that there is a signatory for every signature.
	if len(su.Signatories) != len(su.Signatures) {
		err = hsuerrSignatureSignatoryMismatch
		fmt.Println(err)
		return
	}

	// Check that the Update matches the current block. If it doesn't, it has one
	// step to match the next block.
	if su.Update.Heartbeat.ParentBlock != p.engine.Metadata().ParentBlock {
		time.Sleep(StepDuration)
		if su.Update.Heartbeat.ParentBlock != p.engine.Metadata().ParentBlock {
			err = hsuerrInvalidParent
			fmt.Println(err)
			return
		}
	}

	// Check that there are enough signatures in the update to match the current
	// step.
	p.currentStepLock.Lock()
	if int(p.currentStep) > len(su.Signatures) {
		p.currentStepLock.Unlock()
		err = hsuerrOutOfSync
		fmt.Println(err)
		return
	}
	p.currentStepLock.Unlock()

	// Check that all of the signatures are valid, and that there are no repeats.
	updateHash, err := siacrypto.HashObject(su.Update)
	if err != nil {
		fmt.Println(err)
		return
	}
	message := updateHash[:]
	previousSignatories := make(map[byte]bool)
	for i, signatory := range su.Signatories {
		// Check bounds on current signatory.
		if signatory >= state.QuorumSize {
			err = hsuerrBounds
			fmt.Println(err)
			return
		}

		// Check that current signatory is a valid sibling in the quorum.
		if !p.engine.Metadata().Siblings[signatory].Active {
			err = hsuerrNonSibling
			fmt.Println(err)
			return
		}

		// Check that current signatory has only been seen once in the current SignedUpdate
		if previousSignatories[signatory] {
			err = hsuerrDoubleSign
			fmt.Println(err)
			return
		}
		previousSignatories[signatory] = true

		// Verify the signature.
		verification := p.engine.Metadata().Siblings[signatory].PublicKey.Verify(su.Signatures[i], message)
		if !verification {
			err = hsuerrInvalidSignature
			fmt.Println(err)
			return
		}

		// Extend the signed message so that it contians the proper message for the
		// next verification.
		message = append(su.Signatures[i][:], message...)
	}

	// Check if this update has already been received.
	_, exists := p.updates[su.Signatories[0]][updateHash]
	if exists {
		err = hsuerrHaveHeartbeat
		// no printing because this will happen a lot
		return
	}

	// Check that there are less than two heartbeats from this host yet seen.
	if len(p.updates[su.Signatories[0]]) >= 2 {
		err = hsuerrManyHeartbeats
		fmt.Println(err)
		return
	}

	// Add the update to the list of seen updates.
	p.updates[su.Signatories[0]][updateHash] = su.Update

	// Sign the stack of signatures and append the signature to the stack, then
	// announce the Update to everyone on the quorum
	var signature siacrypto.Signature
	signature, err = p.secretKey.Sign(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	su.Signatures = append(su.Signatures, signature)
	su.Signatories = append(su.Signatories, p.siblingIndex)

	// broadcast the update to the quorum
	return
}
