package consensus

import (
	"delta"
	"errors"
	"fmt"
	"siacrypto"
	"state"
	"time"
)

const (
	StepDuration = 800 * time.Millisecond
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

// The series of printlns in this function are purely for debugging.
func (p *Participant) HandleSignedUpdate(su SignedUpdate, _ *struct{}) (err error) {
	// Lock the engine for the duration of the function.
	p.engineLock.Lock()
	defer p.engineLock.Unlock()

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
	var signedMessage siacrypto.SignedMessage // grows each iteration, as signatures are stacked upon eachother.
	signedMessage.Message = updateHash[:]
	previousSignatories := make(map[byte]bool)
	for i, signatory := range su.Signatories {
		// Check bounds on current signatory.
		if signatory >= state.QuorumSize {
			err = hsuerrBounds
			fmt.Println(err)
			return
		}

		// Check that current signatory is a valid sibling in the quorum.
		if p.engine.Metadata().Siblings[signatory] == nil {
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
		signedMessage.Signature = su.Signatures[i]
		verification := p.engine.Metadata().Siblings[signatory].PublicKey.Verify(signedMessage)
		if !verification {
			err = hsuerrInvalidSignature
			fmt.Println(err)
			return
		}

		// Extend the signed message so that it contians the proper message for the
		// next verification.
		signedMessage.Message = signedMessage.CombinedMessage()
	}

	// Check if this heartbeat has already been received.
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

	// Add the heartbeat to the list of seen heartbeats.
	p.updates[su.Signatories[0]][updateHash] = su.Update

	// Sign the stack of signatures and append the signature to the stack, then announce the Update to everyone on the quorum
	signedMessage, err = p.secretKey.Sign(signedMessage.Message)
	if err != nil {
		fmt.Println(err)
		return
	}
	su.Signatures = append(su.Signatures, signedMessage.Signature)
	su.Signatories = append(su.Signatories, p.self.Index)

	// broadcast the update to the quorum
	return
}

// condenseBlock assumes that a heartbeat has a valid signature and that the
// parent is the correct parent.
func (p *Participant) condenseBlock() (b *delta.Block) {
	/*b = new(delta.Block)
	b.Height = p.quorum.Height()
	b.Parent = p.quorum.Parent()

	p.heartbeatsLock.Lock()
	for i := range p.heartbeats {
		fmt.Printf("Sibling %v: %v heartbeats\n", i, len(p.heartbeats[i]))
		if len(p.heartbeats[i]) == 1 {
			// the map has only one element, but the key is unknown
			for _, hb := range p.heartbeats[i] {
				b.heartbeats[i] = hb // place heartbeat into block, if valid
			}
		}
		p.heartbeats[i] = make(map[siacrypto.Hash]*heartbeat) // clear map for next cycle
	}
	p.heartbeatsLock.Unlock()*/
	return
}
