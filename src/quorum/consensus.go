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
type heartbeat struct {
	entropy  common.Entropy
	hopefuls []*Sibling
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
	err = encoder.Encode(hb.hopefuls)
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
		hb = new(heartbeat)
	}

	r := bytes.NewBuffer(gobHeartbeat)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&hb.entropy)
	if err != nil {
		return
	}
	err = decoder.Decode(&hb.hopefuls)
	return
}

// Using the current State, newSignedHeartbeat() creates a heartbeat for the
// quorum and then signs and announces it.
func (p *Participant) newSignedHeartbeat() (sh *SignedHeartbeat, err error) {
	hb := new(heartbeat)

	// Generate Entropy
	entropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		return
	}
	copy(hb.entropy[:], entropy)

	// incorporate currHeartbeat
	p.currHeartbeatLock.Lock()
	hb.hopefuls = p.currHeartbeat.hopefuls
	p.currHeartbeatLock.Unlock()
	if len(hb.hopefuls) > 0 {
		fmt.Println("including", len(hb.hopefuls), "hopeful(s) in heartbeat")
	}

	sh = new(SignedHeartbeat)

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
	signedHb, err := p.secretKey.Sign(sh.heartbeatHash[:])
	if err != nil {
		return
	}
	sh.signatures[0] = signedHb.Signature
	sh.signatories = make([]byte, 1)
	sh.signatories[0] = p.self.index

	err = p.announceSignedHeartbeat(sh)
	if err != nil {
		log.Fatalln(err)
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
func (p *Participant) HandleSignedHeartbeat(sh SignedHeartbeat, arb *struct{}) error {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.signatures) != len(sh.signatories) {
		return hsherrMismatchedSignatures
	}

	// check that there are not too many signatures and signatories
	if len(sh.signatories) > common.QuorumSize {
		return hsherrOversigned
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
			return hsherrNoSync
		}
	}

	// Check bounds on first signatory
	if int(sh.signatories[0]) >= common.QuorumSize {
		return hsherrBounds
	}

	// we are starting to read from memory, initiate locks
	p.quorum.siblingsLock.RLock()
	p.heartbeatsLock.Lock()
	defer p.quorum.siblingsLock.RUnlock()
	defer p.heartbeatsLock.Unlock()

	// check that first signatory is a sibling
	if p.quorum.siblings[sh.signatories[0]] == nil {
		return hsherrNonSibling
	}

	// Check if we have already received this heartbeat
	_, exists := p.heartbeats[sh.signatories[0]][sh.heartbeatHash]
	if exists {
		return hsherrHaveHeartbeat
	}

	// Check if we already have two heartbeats from this host
	if len(p.heartbeats[sh.signatories[0]]) >= 2 {
		return hsherrManyHeartbeats
	}

	// iterate through the signatures and make sure each is legal
	var signedMessage crypto.SignedMessage // grows each iteration
	signedMessage.Message = sh.heartbeatHash[:]
	previousSignatories := make(map[byte]bool) // which signatories have already signed
	for i, signatory := range sh.signatories {
		// Check bounds on the signatory
		if int(signatory) >= common.QuorumSize {
			return hsherrBounds
		}

		// Verify that the signatory is a sibling in the quorum
		if p.quorum.siblings[signatory] == nil {
			return hsherrNonSibling
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			return hsherrDoubleSigned
		}

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		signedMessage.Signature = sh.signatures[i]
		verification := p.quorum.siblings[signatory].publicKey.Verify(&signedMessage)

		// check status of verification
		if !verification {
			return hsherrInvalidSignature
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
	signedMessage, err := p.secretKey.Sign(signedMessage.Message)
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
	var newSiblings []*Sibling
	// fetch a sibling ordering
	siblingOrdering := p.quorum.siblingOrdering()

	// Lock down s.siblings and s.heartbeats for editing
	p.quorum.siblingsLock.Lock()
	p.heartbeatsLock.Lock()

	var newSeed common.Entropy
	// Read heartbeats, process them, then archive them.
	for _, i := range siblingOrdering {
		// each sibling must submit exactly 1 heartbeat
		if len(p.heartbeats[i]) != 1 {
			p.tossSibling(i)
			continue
		}

		// this is the only way I know to access the only element of a map;
		// the key is unknown
		fmt.Println("Confirming Sibling", i)
		for _, hb := range p.heartbeats[i] {
			newSiblings, newSeed, _ = p.quorum.processHeartbeat(hb, newSeed)
		}

		// archive heartbeats (unimplemented)

		// clear heartbeat list for next block
		p.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
	}

	// add new siblings
	for _, s := range newSiblings {
		err := p.addNewSibling(s)
		if err != nil {
			log.Fatalln("failed to add new sibling:", err)
		}
	}
	if len(newSiblings) != 0 {
		fmt.Println("sending quorum state:")
		fmt.Print(p.quorum.Status())
	}
	// send each new sibling the current quorum state
	for _, s := range newSiblings {
		gobQuorum, _ := p.quorum.GobEncode()
		p.messageRouter.SendAsyncMessage(&common.Message{
			Dest: s.address,
			Proc: "Participant.TransferQuorum",
			Args: gobQuorum,
			Resp: nil,
		})
	}

	p.quorum.siblingsLock.Unlock()
	p.heartbeatsLock.Unlock()

	// move UpcomingEntropy to CurrentEntropy
	p.quorum.seed = newSeed

	// generate, sign, and announce new heartbeat
	_, err := p.newSignedHeartbeat()
	if err != nil {
		return
	}

	// reset the WIP heartbeat
	p.currHeartbeat = heartbeat{}

	return
}

// Tick() updates s.CurrentStep, and calls compile() when all steps are complete
func (p *Participant) tick() {
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
