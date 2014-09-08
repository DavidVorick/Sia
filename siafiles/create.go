package siafiles

import (
	"encoding/base64"
	"errors"
	"os"
	"os/user"
	"path"
)

var (
	siahome   string
	homefound bool

	tempdir string
)

// Homedir returns the users home directory.
func HomeDir() (homedir string, err error) {
	if !homefound {
		err = errors.New("no known home directory")
		return
	}
	homedir = siahome

	return
}

// TempDir returns the Sia temp directory.
func TempDir() string {
	return tempdir
}

// siafiles.init scans the directory structure of Sia, and creates necessary
// folders.
func init() {
	// Set the tempdir.
	tempdir = path.Join(os.TempDir(), "Sia")

	// Create a temp directory for Sia.
	if err := os.MkdirAll(TempDir(), os.ModeDir|os.ModePerm); err != nil {
		// Perhaps something more gentle can be done?
		panic(err)
	}

	// Scan for a home folder.
	homefound = false
	usr, err := user.Current()
	if err != nil {
		// Perhaps a more gentle action can be performed?
		panic(err)
	}
	stdHome := path.Join(usr.HomeDir, ".config", "Sia")
	if Exists(stdHome) {
		siahome = stdHome
		homefound = true
	} else {
		badHome := path.Join(usr.HomeDir, ".Sia")
		if Exists(badHome) {
			siahome = badHome
			homefound = true
		}
	}
}

// HomeFilename returns a safe concetenation of the input filename and the home
// directory.
func HomeFilename(name string) (fullpath string, err error) {
	homedir, err := HomeDir()
	if err != nil {
		return
	}
	fullpath = path.Join(homedir, name)
	return
}

// SafeFilename takes a byteslice and converts it to a string that can be used
// as a filename without casuing errors/unexpected behavior within the
// filesystem.
func SafeFilename(b []byte) (s string) {
	return base64.URLEncoding.EncodeToString(b)
}

// TempFilename returns a safe concatenation of the input filename and the temp
// directory.
func TempFilename(name string) string {
	return path.Join(TempDir(), name)
}
