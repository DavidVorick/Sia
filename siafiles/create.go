package siafiles

import (
	"encoding/base64"
	"os"
	"path"
)

// TempDir is the folder where Sia-specific temporary files, such as
// files created during testing, are stored.
var TempDir string

func init() {
	// initialize the temp directory
	TempDir = os.TempDir() + "/Sia"
	if err := os.MkdirAll(TempDir, os.ModeDir|os.ModePerm); err != nil {
		panic(err)
	}
}

// TempFilename returns a safe concatenation of the input filename
// and the temp directory.
func TempFilename(name string) string {
	return path.Join(TempDir, name)
}

func SafeFilename(b []byte) (s string) {
	return base64.URLEncoding.EncodeToString(b)
}
