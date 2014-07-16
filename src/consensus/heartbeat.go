package consensus

/* import (
	"fmt"
	"quorum"
	"quorum/script"
	"siaencoding"
)

const (
	MinHeartbeatSize = quorum.EntropyVolume + 4
)

// A heartbeat is the object sent by siblings to engage in consensus.
// Heartbeats contain keepalive information as well as a set of scripts
// submitted by arbitrary sources.
type heartbeat struct {
	entropy            quorum.Entropy
	uploadAdvancements []quorum.UploadAdvancement
	scriptInputs       []script.ScriptInput
}

func (hb *heartbeat) GobEncode() (gobHB []byte, err error) {
	if hb == nil {
		err = fmt.Errorf("Cannot encode a nil heartbeat")
		return
	}

	// calculate the size of the encoded heartbeat
	encodedHeartbeatLen := quorum.EntropyVolume + 4 + 4
	encodedHeartbeatLen += quorum.UploadAdvancementSize * len(hb.uploadAdvancements)
	for _, scriptInput := range hb.scriptInputs {
		encodedHeartbeatLen += 12
		encodedHeartbeatLen += len(scriptInput.Input)
	}
	gobHB = make([]byte, encodedHeartbeatLen)

	// copy the entropy over
	copy(gobHB, hb.entropy[:])
	offset := quorum.EntropyVolume

	// copy in the number of uploadAdvancements
	intb := siaencoding.EncUint32(uint32(len(hb.uploadAdvancements)))
	copy(gobHB[offset:], intb[:])
	offset += 4

	// copy each uploadAdvancement
	for i := range hb.uploadAdvancements {
		gobAdvancement, err := hb.uploadAdvancements[i].GobEncode()
		if err != nil {
			panic(err)
		}
		copy(gobHB[offset:], gobAdvancement)
		offset += quorum.UploadAdvancementSize
	}

	// copy in the number of ScriptInputs
	intb = siaencoding.EncUint32(uint32(len(hb.scriptInputs)))
	copy(gobHB[offset:], intb[:])
	offset += 4

	// copy in each scriptInput, while also copying in the offset for each
	// scriptInput
	scriptInputOffset := uint32(offset + len(hb.scriptInputs)*4)
	for _, scriptInput := range hb.scriptInputs {
		// copy over the offset
		intb := siaencoding.EncUint32(scriptInputOffset)
		copy(gobHB[offset:], intb[:])
		offset += 4

		// copy over the ScriptInput
		id := scriptInput.WalletID.Bytes()
		copy(gobHB[scriptInputOffset:], id[:])
		scriptInputOffset += quorum.WalletIDSize
		n := copy(gobHB[scriptInputOffset:], scriptInput.Input)
		scriptInputOffset += uint32(n)
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
	if len(gobHB) < quorum.EntropyVolume+4+4 {
		err = fmt.Errorf("Received invalid encoded heartbeat")
		return
	}

	// copy over the entropy
	copy(hb.entropy[:], gobHB)
	offset := quorum.EntropyVolume

	// read the number of uploadAdvancements
	numUploadAdvancements := siaencoding.DecUint32(gobHB[offset : offset+4])
	offset += 4
	if len(gobHB) < quorum.EntropyVolume+4+4+(int(numUploadAdvancements)*quorum.UploadAdvancementSize) {
		err = fmt.Errorf("Received invalid encoded heartbeat")
		return
	}
	hb.uploadAdvancements = make([]quorum.UploadAdvancement, numUploadAdvancements)
	for i := range hb.uploadAdvancements {
		hb.uploadAdvancements[i].GobDecode(gobHB[offset : offset+quorum.UploadAdvancementSize])
	}

	// get the number of ScriptInputs
	numScriptInputs := int(siaencoding.DecUint32(gobHB[offset : offset+4]))
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
	nextOffset := uint32(offset + 4)
	hb.scriptInputs = make([]script.ScriptInput, numScriptInputs)
	for i := 0; i < int(numScriptInputs-1); i++ {
		siOffset := siaencoding.DecUint32(gobHB[offset : offset+4])
		nextOffset = siaencoding.DecUint32(gobHB[offset+4 : offset+8])

		if siOffset > nextOffset-quorum.WalletIDSize || nextOffset+quorum.WalletIDSize > uint32(len(gobHB)) {
			err = fmt.Errorf("Received invalid encoded heartbeat")
			return
		}

		// decode the WalletID
		hb.scriptInputs[i].WalletID = quorum.WalletID(siaencoding.DecUint64(gobHB[siOffset : siOffset+8]))
		siOffset += quorum.WalletIDSize
		hb.scriptInputs[i].Input = gobHB[siOffset:nextOffset]

		offset += 4
	}

	hb.scriptInputs[numScriptInputs-1].WalletID = quorum.WalletID(siaencoding.DecUint64(gobHB[nextOffset : nextOffset+8]))
	nextOffset += quorum.WalletIDSize
	hb.scriptInputs[numScriptInputs-1].Input = gobHB[nextOffset:]

	return
} */
