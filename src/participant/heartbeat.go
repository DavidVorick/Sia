package participant

import (
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

func (hb *heartbeat) GobEncode() (gobHB []byte, err error) {
	if hb == nil {
		err = fmt.Errorf("Cannot encode a nil heartbeat")
		return
	}

	// calculate the size of the encoded heartbeat
	encodedHeartbeatLen := quorum.EntropyVolume + 4
	for _, scriptInput := range hb.scriptInputs {
		encodedHeartbeatLen += 12
		encodedHeartbeatLen += len(scriptInput.Input)
	}
	gobHB = make([]byte, encodedHeartbeatLen)

	// copy the entropy over
	copy(gobHB, hb.entropy[:])
	offset := quorum.EntropyVolume

	// copy in the number of ScriptInputs
	intb := siaencoding.IntToByte(len(hb.scriptInputs))
	copy(gobHB[offset:], intb[:])
	offset += 4

	// copy in each scriptInput, while also copying in the offset for each
	// scriptInput
	scriptInputOffset := offset + len(hb.scriptInputs)*4
	for _, scriptInput := range hb.scriptInputs {
		// copy over the offset
		intb := siaencoding.IntToByte(scriptInputOffset)
		copy(gobHB[offset:], intb[:])
		offset += 4

		// copy over the ScriptInput
		id := scriptInput.WalletID.Bytes()
		copy(gobHB[scriptInputOffset:], id[:])
		scriptInputOffset += quorum.WalletIDSize
		n := copy(gobHB[scriptInputOffset:], scriptInput.Input)
		scriptInputOffset += n
	}

	return
}

func (hb *heartbeat) GobDecode(gobHB []byte) (err error) {
	// check for a nil heartbeat
	if hb == nil {
		err = fmt.Errorf("Cannot decode into nil heartbeat")
		return
	}
	// check for a too-short byte slice
	if len(gobHB) < quorum.EntropyVolume+4 {
		err = fmt.Errorf("Received invalid encoded heartbeat")
		return
	}

	// copy over the entropy
	copy(hb.entropy[:], gobHB)
	offset := quorum.EntropyVolume

	// get the number of ScriptInputs
	println(len(gobHB))
	println(offset)
	var intb [4]byte
	copy(intb[:], gobHB[offset:])
	numScriptInputs := siaencoding.IntFromByte(intb)
	if numScriptInputs == 0 {
		return
	}
	offset += 4

	// make sure there are at least enough bytes for all the offsets
	if len(gobHB)-quorum.WalletIDSize < offset+4*numScriptInputs {
		err = fmt.Errorf("Received invalid encoded heartbeat")
		return
	}

	// decode each script input
	var nextOffset int
	var uint64b [8]byte
	hb.scriptInputs = make([]script.ScriptInput, numScriptInputs)
	for i := 0; i < numScriptInputs-1; i++ {
		copy(intb[:], gobHB[offset:])
		siOffset := siaencoding.IntFromByte(intb)
		copy(intb[:], gobHB[offset+4:])
		nextOffset = siaencoding.IntFromByte(intb)

		if siOffset > nextOffset-quorum.WalletIDSize || nextOffset+quorum.WalletIDSize > len(gobHB) {
			err = fmt.Errorf("Received invalid encoded heartbeat")
			return
		}

		// decode the WalletID
		copy(uint64b[:], gobHB[siOffset:])
		hb.scriptInputs[i].WalletID = quorum.WalletID(siaencoding.UInt64FromByte(uint64b))
		siOffset += quorum.WalletIDSize
		hb.scriptInputs[i].Input = gobHB[siOffset:nextOffset]

		offset += 4
	}

	copy(uint64b[:], gobHB[nextOffset:])
	hb.scriptInputs[numScriptInputs-1].WalletID = quorum.WalletID(siaencoding.UInt64FromByte(uint64b))
	nextOffset += quorum.WalletIDSize
	hb.scriptInputs[numScriptInputs-1].Input = gobHB[nextOffset:]

	return
}
