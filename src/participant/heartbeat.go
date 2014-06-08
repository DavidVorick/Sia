package participant

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"quorum"
	"quorum/script"
	"siaencoding"
)

// A heartbeat is the object sent by siblings to engage in consensus.
// Heartbeats contain keepalive information as well as a set of scripts
// submitted by arbitrary sources.
type heartbeat struct {
	entropy      quorum.Entropy
	scriptInputs []script.ScriptInput
}

func (hb *heartbeat) GobEncode() (gobHeartbeat []byte, err error) {
	if hb == nil {
		err = fmt.Errorf("Cannot encode a nil heartbeat")
	}

	// calculate the size of the encoded heartbeat
	encodedHeartbeatLen := quorum.EntropyVolume + 4
	for i, scriptInput := range hb.scriptInputs {
		encodedHeartbeatLen += 12
		encodedHeartbeatLen += len(scriptInput.Input)
	}
	gobHeartbeat = make([]byte, encodedHeartbeatLen)

	// copy the entropy over
	copy(gobHeartbeat, hb.entropy[:])
	offset := quorum.EntropyVolume

	// copy in the number of ScriptInputs
	intb := siaencoding.IntToByte(len(hb.scriptInputs))
	copy(gobHeartbeat[offset:], intb[:])
	offset += 4

	// copy in each scriptInput, while also copying in the offset for each
	// scriptInput
	scriptInputOffset := offset + len(hb.scriptInputs)*4
	for i, scriptInput := range hb.scriptInputs {
		// copy over the offset
		intb := siaencoding.IntToByte(scriptInputOffset)
		copy(gobHeartbeat[offset:], intb[:])
		offset += 4

		// copy over the ScriptInput
		id := scriptInput.WalletID.Bytes()
		copy(gobHeartbeat[scriptInputOffset:], id[:])
		scriptInputOffset += quorum.WalletIDSize
		n := copy(gobHeartbeat[scriptInputOffset:], scriptInput.Input)
		scriptInputOffset += n
	}

	return
}

func (hb *heartbeat) GobDecode(gobHeartbeat []byte) (err error) {
	if hb == nil {
		err = fmt.Errorf("Cannot decode into nil heartbeat")
		return
	}

	return
}
