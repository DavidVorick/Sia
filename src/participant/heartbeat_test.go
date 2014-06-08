package participant

import (
	"quorum"
	"reflect"
	"siacrypto"
	"testing"
)

func TestHeartbeatEncoding(t *testing.T) {
	// encode a nil heartbeat
	var hb *heartbeat
	ehb, err := hb.GobEncode()
	if err == nil {
		t.Fatal("no error produced when encoding a nil heartbeat")
	}

	// create entropy for the heartbeat
	hb = new(heartbeat)
	entropy, err := siacrypto.RandomByteSlice(quorum.EntropyVolume)
	if err != nil {
		t.Fatal(err)
	}
	copy(hb.entropy[:], entropy)

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

	// reflect.DeepEqual checks each value, including for the maps
	equal := reflect.DeepEqual(hb, dhb)
	if !equal {
		t.Error("heartbeat not identical after being encoded then decoded")
	}
}
