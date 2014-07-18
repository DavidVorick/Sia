package delta

import (
	"siacrypto"
	"state"
)

func (e *Engine) Compile(b Block) {
	// The first thing that happens is the entropy seed for the block is
	// determined. Though not implemented, this happens by pulling the latest
	// external entropy source from the block and hashing it against the germ
	// from the previous block. Right now, the germ is created but the portion
	// about external entropy is omitted.
	var externalEntropy state.Entropy // will be pulled from block
	e.state.MergeExternalEntropy(externalEntropy)

	// Next each heartbeat is iterated through and processed, checking that all
	// the vital information has been correctly assembled.
	var siblingEntropy []byte
	for i, heartbeat := range b.Heartbeats {
		// Ignore heartbeat if there's no sibling.
		if !e.state.Metadata.Siblings[i].Active {
			continue
		}

		// Verify the signature on the heartbeat.
		heartbeatBytes, err := siacrypto.HashObject(heartbeat)
		if err != nil {
			e.state.TossSibling(byte(i))
			continue
		}
		signedHeartbeatBytes := siacrypto.SignedMessage{
			Signature: b.HeartbeatSignatures[i],
			Message:   heartbeatBytes[:],
		}
		if !e.state.Metadata.Siblings[i].PublicKey.Verify(signedHeartbeatBytes) {
			e.state.TossSibling(byte(i))
			continue
		}

		// Verify the parent block of the heartbeat.
		if heartbeat.ParentBlock != e.state.Metadata.ParentBlock {
			e.state.TossSibling(byte(i))
			continue
		}

		// Append the entropy to siblingEntropy.
		siblingEntropy = append(siblingEntropy, heartbeat.Entropy[:]...)
	}
}
