package siafiles

import (
	"encoding/base64"
	"strings"
)

func SafeFilename(b []byte) (s string) {
	if b == nil {
		return
	}

	s = base64.StdEncoding.EncodeToString(b)
	s = strings.Replace(s, "/", "_", -1)
	return
}
