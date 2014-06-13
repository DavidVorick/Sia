package siaencoding

import (
	"encoding/base64"
	"strings"
)

func EncFilename(b []byte) (s string) {
	s = base64.StdEncoding.EncodeToString(b)
	s = strings.Replace(s, "/", "_", -1)
	return
}
