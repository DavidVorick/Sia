package siaencoding

import (
	"encoding/json"
)

// Marshal is the generic function used to marshal all public and over-the-wire
// data used on Sia. By implementation, it's just making a call to json. We use
// our own function however because if the spec is changed to use an encoding
// other than json (which is likely), then it we will only need to make changes
// in one place.
func Marshal(v interface{}) (b []byte, err error) {
	b, err = json.Marshal(v)
	return
}
