package consensus

/* import (
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
}*/
