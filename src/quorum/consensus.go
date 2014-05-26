package quorum

import (
	"bytes"
	"common"
	"common/crypto"
	"common/log"
	"encoding/gob"
	"errors"
	"fmt"
	"time"
)

// All information that needs to be passed between siblings each block
// This may need to be changed to a slice of updates, or a slice of []bytes (encoded updates)
// The participant should have a map of updates. Upon copying the map into the heartbeat struct,
// the participant can do a 'for range' if that makes more sense. I'm not particularly attatched
// to map[Update]Update in the heartbeat itself. Change it to whatever makes the most sense to you.
// Thanks!
type heartbeat struct {
	entropy common.Entropy
	updates []Update
}

// Contains a heartbeat that has been signed iteratively, is a key part of the
// signed solution to the Byzantine Generals Problem
type SignedHeartbeat struct {
	heartbeat     *heartbeat
	heartbeatHash crypto.TruncatedHash
	signatories   []byte             // a list of everyone who's seen the heartbeat
	signatures    []crypto.Signature // their corresponding signatures
}

// Convert heartbeat to []byte
func (hb *heartbeat) GobEncode() (gobHeartbeat []byte, err error) {
	// if hb == nil, encode a zero heartbeat
	if hb == nil {
		hb = new(heartbeat)
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(hb.entropy)
	if err != nil {
		return
	}
	err = encoder.Encode(hb.updates)
	if err != nil {
		return
	}

	gobHeartbeat = w.Bytes()
	return
}

// Convert []byte to heartbeat
func (hb *heartbeat) GobDecode(gobHeartbeat []byte) (err error) {
	// if hb == nil, make a new heartbeat and decode into that
	if hb == nil {
		err = fmt.Errorf("Cannot decode into nil heartbeat")
		return
	}

	r := bytes.NewBuffer(gobHeartbeat)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&hb.entropy)
	if err != nil {
		return
	}
	err = decoder.Decode(&hb.updates)
	if err != nil {
		return
	}
	return
}

// Takes a signed heartbeat and broadcasts it to the quorum
func (p *Participant) announceSignedHeartbeat(sh *SignedHeartbeat) (err error) {
	p.broadcast(&common.Message{
		Proc: "Participant.HandleSignedHeartbeat",
		Args: *sh,
		Resp: nil,
	})
	return
}

// Using the current State, newSignedHeartbeat() creates a heartbeat for the
// quorum and then signs and announces it.
func (p *Participant) newSignedHeartbeat() (err error) {
	hb := new(heartbeat)

	// Generate Entropy
	entropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		return
	}
	copy(hb.entropy[:], entropy)

	// Add updates gathered since last compile and clear the list
	p.updatesLock.Lock()
	for _, value := range p.updates {
		hb.updates = append(hb.updates, value)
	}
	p.updates = make(map[Update]Update)
	p.updatesLock.Unlock()

	sh := new(SignedHeartbeat)

	// confirm heartbeat and hash
	sh.heartbeat = hb
	gobHb, err := hb.GobEncode()
	if err != nil {
		return
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(gobHb)
	if err != nil {
		return
	}

	// fill out signatures
	sh.signatures = make([]crypto.Signature, 1)
	hbSignature, err := p.secretKey.Sign(sh.heartbeatHash[:])
	if err != nil {
		return
	}
	sh.signatures[0] = hbSignature.Signature
	sh.signatories = make([]byte, 1)
	sh.signatories[0] = p.self.index

	err = p.announceSignedHeartbeat(sh)
	if err != nil {
		log.Fatalln(err)
	}
	return
}

var hsherrMismatchedSignatures = errors.New("SignedHeartbeat has mismatches signatures to signatories")
var hsherrOversigned = errors.New("Received an over-signed signedHeartbeat")
var hsherrNoSync = errors.New("Received an out-of-sync SignedHeartbeat")
var hsherrBounds = errors.New("Received an out of bounds index for signatory")
var hsherrNonSibling = errors.New("Received heartbeat from non-sibling")
var hsherrHaveHeartbeat = errors.New("Already have this heartbeat")
var hsherrManyHeartbeats = errors.New("Received many heartbeats from this host")
var hsherrDoubleSigned = errors.New("Received a double signature")
var hsherrInvalidSignature = errors.New("Received heartbeat with invalid signature")

// HandleSignedHeartbeat takes the payload of an incoming message of type
// 'incomingSignedHeartbeat' and verifies it according to the specification
//
// What sort of input error checking is needed for this function?
func (p *Participant) HandleSignedHeartbeat(sh SignedHeartbeat, arb *struct{}) (err error) {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.signatures) != len(sh.signatories) {
		err = hsherrMismatchedSignatures
		return
	}

	// check that there are not too many signatures and signatories
	if len(sh.signatories) > common.QuorumSize {
		err = hsherrOversigned
		return
	}

	p.stepLock.Lock() // prevents a benign race condition; is here to follow best practices
	currentStep := p.currentStep
	p.stepLock.Unlock()
	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless
	// there is a new block and s.CurrentStep is common.QuorumSize
	if currentStep > len(sh.signatories) {
		if currentStep == common.QuorumSize && len(sh.signatories) == 1 {
			// by waiting common.StepDuration, the new block will be compiled
			time.Sleep(common.StepDuration)
			// now continue to rest of function
		} else {
			err = hsherrNoSync
			return
		}
	}

	// Check bounds on first signatory
	if int(sh.signatories[0]) >= common.QuorumSize {
		err = hsherrBounds
		return
	}

	// we are starting to read from memory, initiate locks
	p.quorum.lock.RLock()
	p.heartbeatsLock.Lock()
	defer p.quorum.lock.RUnlock()
	defer p.heartbeatsLock.Unlock()

	// check that first signatory is a sibling
	if p.quorum.siblings[sh.signatories[0]] == nil {
		err = hsherrNonSibling
		return
	}

	// Check if we have already received this heartbeat
	_, exists := p.heartbeats[sh.signatories[0]][sh.heartbeatHash]
	if exists {
		err = hsherrHaveHeartbeat
		return
	}

	// Check if we already have two heartbeats from this host
	if len(p.heartbeats[sh.signatories[0]]) >= 2 {
		err = hsherrManyHeartbeats
		return
	}

	// iterate through the signatures and make sure each is legal
	var signedMessage crypto.SignedMessage // grows each iteration
	signedMessage.Message = sh.heartbeatHash[:]
	previousSignatories := make(map[byte]bool) // which signatories have already signed
	for i, signatory := range sh.signatories {
		// Check bounds on the signatory
		if int(signatory) >= common.QuorumSize {
			err = hsherrBounds
			return
		}

		// Verify that the signatory is a sibling in the quorum
		if p.quorum.siblings[signatory] == nil {
			err = hsherrNonSibling
			return
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			err = hsherrDoubleSigned
			return
		}

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		signedMessage.Signature = sh.signatures[i]
		verification := p.quorum.siblings[signatory].publicKey.Verify(&signedMessage)

		// check status of verification
		if !verification {
			err = hsherrInvalidSignature
			return
		}

		// throwing the signature into the message here makes code cleaner in the loop
		// and after we sign it to send it to everyone else
		newMessage, err := signedMessage.CombinedMessage()
		signedMessage.Message = newMessage
		if err != nil {
			return err
		}
	}

	// Add heartbeat to list of seen heartbeats
	p.heartbeats[sh.signatories[0]][sh.heartbeatHash] = sh.heartbeat

	// Sign the stack of signatures and send it to all hosts
	signedMessage, err = p.secretKey.Sign(signedMessage.Message)
	if err != nil {
		log.Fatalln(err)
	}

	// add our signature to the signedHeartbeat
	sh.signatures = append(sh.signatures, signedMessage.Signature)
	sh.signatories = append(sh.signatories, p.self.index)

	// broadcast the message to the quorum
	err = p.announceSignedHeartbeat(&sh)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	return nil
}

func (sh *SignedHeartbeat) GobEncode() (gobSignedHeartbeat []byte, err error) {
	// error check the input
	if sh == nil {
		err = fmt.Errorf("Cannot encode a nil object")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(sh.heartbeat)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.heartbeatHash)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.signatories)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.signatures)
	if err != nil {
		return
	}

	gobSignedHeartbeat = w.Bytes()
	return
}

func (sh *SignedHeartbeat) GobDecode(gobSignedHeartbeat []byte) (err error) {
	if gobSignedHeartbeat == nil {
		err = fmt.Errorf("cannot decode a nil byte slice")
		return
	}

	r := bytes.NewBuffer(gobSignedHeartbeat)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&sh.heartbeat)
	if err != nil {
		return
	}
	err = decoder.Decode(&sh.heartbeatHash)
	if err != nil {
		return
	}
	err = decoder.Decode(&sh.signatories)
	if err != nil {
		return
	}
	err = decoder.Decode(&sh.signatures)
	if err != nil {
		return
	}

	return
}

// compile() takes the list of heartbeats and uses them to advance the state.
//
// Needs updated error handling
func (p *Participant) compile() {
	// Lock down s.heartbeats and quorum for editing
	p.quorum.lock.Lock()
	p.heartbeatsLock.Lock()

	// fetch a sibling ordering
	siblingOrdering := p.quorum.siblingOrdering()

	// Read heartbeats, process them, then archive them.
	var newSeed common.Entropy
	updateList := make(map[Update]bool)
	for _, i := range siblingOrdering {
		// each sibling must submit exactly 1 heartbeat
		if len(p.heartbeats[i]) != 1 {
			p.quorum.tossSibling(i)
			continue
		}

		// this is the only way I know to access the only element of a map;
		// the key is unknown
		fmt.Println("Confirming Sibling", i)
		for _, hb := range p.heartbeats[i] {
			p.processHeartbeat(hb, &newSeed, updateList)
		}

		// archive heartbeats (tbi)

		// clear heartbeat list for next block
		p.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
	}

	// copy the new seed into the quorum
	p.quorum.seed = newSeed

	// print the status of the quorum after compiling
	fmt.Print(p.quorum.Status())

	p.quorum.lock.Unlock()
	p.heartbeatsLock.Unlock()

	// create new heartbeat (it gets broadcast automatically)
	p.newSignedHeartbeat()
	return
}

// Tick() updates s.CurrentStep, and calls compile() when all steps are complete
func (p *Participant) tick() {
	p.tickingLock.Lock()
	if p.ticking {
		p.tickingLock.Unlock()
		return
	}
	p.ticking = true
	p.tickingLock.Unlock()

	// Every common.StepDuration, advance the state stage
	ticker := time.Tick(common.StepDuration)
	for _ = range ticker {
		p.stepLock.Lock()
		if p.currentStep == common.QuorumSize {
			fmt.Println("compiling")
			p.compile()
			p.currentStep = 1
		} else {
			fmt.Println("stepping")
			p.currentStep += 1
		}
		p.stepLock.Unlock()
	}
}
