package participant

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"logger"
	"network"
	"quorum"
	"siacrypto"
	"time"
)

type SignedHeartbeat struct {
	heartbeat     *heartbeat
	heartbeatHash siacrypto.Hash
	signatories   []byte                // a list of everyone who's seen the heartbeat
	signatures    []siacrypto.Signature // their corresponding signatures
}

// Takes a signed heartbeat and broadcasts it to the quorum
func (p *Participant) announceSignedHeartbeat(sh *SignedHeartbeat) (err error) {
	p.broadcast(&network.Message{
		Proc: "Participant.HandleSignedHeartbeat",
		Args: *sh,
		Resp: err,
	})
	return
}

// Using the current State, newSignedHeartbeat() creates a heartbeat for the
// quorum and then signs and announces it.
func (p *Participant) newSignedHeartbeat() (err error) {
	hb := new(heartbeat)

	// Generate Entropy
	entropy := siacrypto.RandomByteSlice(quorum.EntropyVolume)
	if err != nil {
		return
	}
	copy(hb.entropy[:], entropy)

	// Copy scripts
	p.scriptInputsLock.Lock()
	hb.scriptInputs = p.scriptInputs
	p.scriptInputs = nil
	p.scriptInputsLock.Unlock()
	p.uploadAdvancementsLock.Lock()
	hb.uploadAdvancements = p.uploadAdvancements
	p.uploadAdvancements = nil
	p.uploadAdvancementsLock.Unlock()

	sh := new(SignedHeartbeat)

	// Place heartbeat into signed heartbeat with hash
	sh.heartbeat = hb
	hbb, _ := hb.GobEncode()
	sh.heartbeatHash = siacrypto.CalculateHash(hbb)

	// fill out signatures
	sh.signatures = make([]siacrypto.Signature, 1)
	hbSignature, err := p.secretKey.Sign(sh.heartbeatHash[:])
	if err != nil {
		return
	}
	sh.signatures[0] = hbSignature.Signature
	sh.signatories = make([]byte, 1)
	sh.signatories[0] = p.self.Index()

	// Add heartbeat to list of seen heartbeats and announce it
	p.heartbeats[p.self.Index()][sh.heartbeatHash] = sh.heartbeat
	err = p.announceSignedHeartbeat(sh)
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
func (p *Participant) HandleSignedHeartbeat(sh SignedHeartbeat, _ *struct{}) (err error) {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.signatures) != len(sh.signatories) {
		err = hsherrMismatchedSignatures
		fmt.Println(err)
		return
	}

	// check that there are not too many signatures and signatories
	if len(sh.signatories) > int(quorum.QuorumSize) {
		err = hsherrOversigned
		fmt.Println(err)
		return
	}

	// check that heartbeat matches heartbeatHash
	// tbi

	p.stepLock.Lock() // prevents a benign race condition; is here to follow best practices
	currentStep := p.currentStep
	p.stepLock.Unlock()
	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless
	// there is a new block and s.CurrentStep is QuorumSize
	//
	// IMPORTANT: synchronizaation is broken, and hot-fixed together in an
	// insecure way. What's important is that the parents block line up.
	if currentStep > len(sh.signatories) {
		if currentStep == int(quorum.QuorumSize) {
			// by waiting StepDuration, the new block will be compiled
			time.Sleep(StepDuration)
			// now continue to rest of function
		} else {
			err = hsherrNoSync
			fmt.Println(err)
			return
		}
	}

	// Check bounds on first signatory
	if sh.signatories[0] >= quorum.QuorumSize {
		err = hsherrBounds
		fmt.Println(err)
		return
	}

	// we are starting to read from memory, initiate locks
	p.heartbeatsLock.Lock()
	defer p.heartbeatsLock.Unlock()

	siblings := p.quorum.Siblings()

	// check that first signatory is a sibling
	if siblings[sh.signatories[0]] == nil {
		err = hsherrNonSibling
		fmt.Println(err)
		return
	}

	// Check if we have already received this heartbeat
	_, exists := p.heartbeats[sh.signatories[0]][sh.heartbeatHash]
	if exists {
		err = hsherrHaveHeartbeat
		// fmt.Println(err)
		return
	}

	// Check if we already have two heartbeats from this host
	if len(p.heartbeats[sh.signatories[0]]) >= 2 {
		err = hsherrManyHeartbeats
		fmt.Println(err)
		return
	}

	// iterate through the signatures and make sure each is legal
	var signedMessage siacrypto.SignedMessage // grows each iteration
	signedMessage.Message = sh.heartbeatHash[:]
	previousSignatories := make(map[byte]bool) // which signatories have already signed
	for i, signatory := range sh.signatories {
		// Check bounds on the signatory
		if signatory >= quorum.QuorumSize {
			err = hsherrBounds
			fmt.Println(err)
			return
		}

		// Verify that the signatory is a sibling in the quorum
		if siblings[signatory] == nil {
			err = hsherrNonSibling
			fmt.Println(err)
			return
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			err = hsherrDoubleSigned
			fmt.Println(err)
			return
		}

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		signedMessage.Signature = sh.signatures[i]
		key := siblings[signatory].PublicKey()
		verification := key.Verify(&signedMessage)

		// check status of verification
		if !verification {
			err = hsherrInvalidSignature
			return
		}

		// throwing the signature into the message here makes code cleaner in the loop
		// and after we sign it to send it to everyone else
		newMessage, err := signedMessage.GobEncode()
		signedMessage.Message = newMessage
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Add heartbeat to list of seen heartbeats
	p.heartbeats[sh.signatories[0]][sh.heartbeatHash] = sh.heartbeat

	// Sign the stack of signatures and send it to all hosts
	signedMessage, err = p.secretKey.Sign(signedMessage.Message)
	if err != nil {
		logger.Fatalln(err)
	}

	// add our signature to the signedHeartbeat
	sh.signatures = append(sh.signatures, signedMessage.Signature)
	sh.signatories = append(sh.signatories, p.self.Index())

	// broadcast the message to the quorum
	err = p.announceSignedHeartbeat(&sh)
	if err != nil {
		logger.Fatalln(err)
	}

	return
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
