package quorum

import (
	"common"
	"common/crypto"
	"fmt"
)

type Participant struct {
	// The quorum in which the participant participates
	quorum quorum

	// Variables local to the participant
	messageRouter common.MessageRouter
	self          *Sibling         // the sibling object for this participant
	secretKey     crypto.SecretKey // secret key matching self.publicKey
}

// CreateParticipant creates a participant.
func CreateParticipant(messageRouter common.MessageRouter) (p *Participant, err error) {
	// check for non-nil messageRouter
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// create a signature keypair for this participant
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}

	// initialize State with default values and keypair
	p = &Participant{
		quorum: quorum{
			currentStep: 1,
		},
		messageRouter: messageRouter,
		self: &Sibling{
			index:     255,
			address:   messageRouter.Address(),
			publicKey: pubKey,
		},
		secretKey: secKey,
	}

	// register State and store our assigned ID
	p.self.address.ID = messageRouter.RegisterHandler(p)

	// send join request to bootstrap
	errChan := p.messageRouter.SendAsyncMessage(&common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.JoinSia",
		Args: *p.self,
		Resp: nil,
	})
	err = <-errChan
	return
}

// Takes a Message and broadcasts it to every Sibling in the quorum
func (p *Participant) broadcast(m *common.Message) {
	p.quorum.siblingsLock.RLock()
	for i := range p.quorum.siblings {
		if p.quorum.siblings[i] != nil {
			nm := *m
			nm.Dest = p.quorum.siblings[i].address
			p.messageRouter.SendAsyncMessage(&nm)
		}
	}
	p.quorum.siblingsLock.RUnlock()
}
