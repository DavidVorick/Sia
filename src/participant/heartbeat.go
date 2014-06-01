package participant

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"quorum"
)

type heartbeat struct {
	entropy quorum.Entropy
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
	return
}
