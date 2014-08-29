package siafiles

import (
	"os"
)

// ValidName returns whether a string is a valid filename. The most reliable
// way to do this is by attempting to open or create the file.
func ValidName(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	} else if _, err := os.Create(filename); err == nil {
		os.Remove(filename)
		return true
	}
	return false
}

// Remove is a simple wrapper for os.Remove. However, like os.RemoveAll, it does
// not return an error if the file does not exist.
func Remove(filename string) error {
	if err := os.Remove(filename); !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists returns whether or not a file exists.
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
