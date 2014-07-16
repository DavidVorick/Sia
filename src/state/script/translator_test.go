package script

/* import (
	"bytes"
	"siacrypto"
	"testing"
)

// TestTranslator ensures that a script is unchanged after being translated back and forth
func TestTranslator(t *testing.T) {
	pk := new(siacrypto.PublicKey)
	copy(pk[:], siacrypto.RandomByteSlice(32))
	s := DefaultScript(pk)

	w, err := BytesToWords(s)
	if err != nil {
		t.Fatal(err)
	}

	b, err := WordsToBytes(w)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(s, b) != 0 {
		t.Fatal("scripts do not match after translate/untranslate")
	}
}*/
