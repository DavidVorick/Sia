package script

import (
	"bytes"
	"encoding/gob"
	"network"
	"quorum"
	"siacrypto"
)

type ScriptInput struct {
	// Wallet quorum.WalletID
	Input []byte
}

func (s *ScriptInput) Bytes() (b []byte) {
	b = s.Input
	return
}

func (s *ScriptInput) Interpret(q *quorum.Quorum) {
	var address network.Address
	var key siacrypto.PublicKey
	r := bytes.NewBuffer(s.Input)
	decoder := gob.NewDecoder(r)
	decoder.Decode(&address)
	decoder.Decode(&key)

	sibling := quorum.NewSibling(address, &key)

	q.AddSibling(sibling)
}
