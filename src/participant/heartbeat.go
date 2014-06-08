package participant

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"quorum"
	"quorum/script"
)

// A heartbeat is the object sent by siblings to engage in consensus.
// Heartbeats contain keepalive information as well as a set of scripts
// submitted by arbitrary sources.
type heartbeat struct {
	entropy      quorum.Entropy
	scriptInputs []script.ScriptInput
}

func (hb *heartbeat) Bytes() (b []byte) {
	b = append(b, hb.entropy[:]...)
	for _, script := range hb.scriptInputs {
		b = append(b, script.Bytes()...)
	}
	return
}

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
	err = encoder.Encode(hb.scriptInputs)
	if err != nil {
		return
	}

	gobHeartbeat = w.Bytes()
	return
}

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
	err = decoder.Decode(&hb.scriptInputs)
	return
}
