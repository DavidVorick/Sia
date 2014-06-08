package participant

import (
	"quorum"
	"quorum/script"
	"siacrypto"
	"testing"
)

// In addition to trying a handful of bogus inputs, TestHeartbeatEncoding fills
// out scripts and encodes then decodes heartbeats to see if they still come
// out valid.
func TestHeartbeatEncoding(t *testing.T) {
	// encode a nil heartbeat
	var hb *heartbeat
	ehb, err := hb.GobEncode()
	if err == nil {
		t.Fatal("no error produced when encoding a nil heartbeat")
	}

	// create entropy for the heartbeat
	hb = new(heartbeat)
	entropy := siacrypto.RandomByteSlice(quorum.EntropyVolume)
	copy(hb.entropy[:], entropy)

	// create random scriptInptus for the heartbeat
	hb.scriptInputs = make([]script.ScriptInput, 3)
	hb.scriptInputs[0] = script.ScriptInput{
		WalletID: quorum.WalletID(siacrypto.RandomUInt64()),
		Input:    siacrypto.RandomByteSlice(15),
	}

	// encode the filled out heartbeat
	ehb, err = hb.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// decode into a nil heartbeat
	var dhb *heartbeat
	err = dhb.GobDecode(ehb)
	if err == nil {
		t.Error("heartbeat.GobDecode accepts a nil heartbeat")
	}

	// decode into non-nil heartbeat
	dhb = new(heartbeat)
	err = dhb.GobDecode(ehb)
	if err != nil {
		t.Fatal(err)
	}

	if hb.entropy != dhb.entropy {
		t.Error("heartbeat entropies not the same after being encoded then decoded")
		t.Error(hb.entropy)
		t.Error(dhb.entropy)
	}

	if len(hb.scriptInputs) != len(dhb.scriptInputs) {
		t.Fatal("heartbeat and decoded heartbeat have different scriptInput volumes!")
	}

	for i := range hb.scriptInputs {
		if hb.scriptInputs[i].WalletID != dhb.scriptInputs[i].WalletID {
			t.Error("WalletIDs are not equal!")
		}
		if len(hb.scriptInputs[i].Input) != len(dhb.scriptInputs[i].Input) {
			t.Error("Script Inputs are not identical - lengths are different")
		}

		for j := range hb.scriptInputs[i].Input {
			if hb.scriptInputs[i].Input[j] != dhb.scriptInputs[i].Input[j] {
				t.Fatal("Script inputs are not matching in content")
			}
		}
	}
}
