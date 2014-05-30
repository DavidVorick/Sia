// package crypto uses libsodium and manages all of the crypto
// for Sia. It has an explicit typing system that uses byte
// arrays matching the sizes specified by the libsodium constants.
package siacrypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"math/big"
)

const (
	// sizes in bytes
	HashSize          int = 64
	TruncatedHashSize int = 32
)

// The underlying variables to the keys & signatures should not be exported
type PublicKey struct {
	key ecdsa.PublicKey
}
type SecretKey struct {
	key ecdsa.PrivateKey
}
type Signature struct {
	r, s *big.Int
}

type Hash [HashSize]byte
type TruncatedHash [TruncatedHashSize]byte

// Compare returns true if the keys are composed of the same integer values
// Compare returns false if any sub-value is nil
func (pk0 *PublicKey) Compare(pk1 *PublicKey) bool {
	// return false if either sub-value is nil
	if pk0.key.X == nil || pk0.key.Y == nil || pk1.key.X == nil || pk1.key.Y == nil {
		return false
	}

	cmp := pk0.key.X.Cmp(pk1.key.X)
	if cmp != 0 {
		return false
	}

	cmp = pk0.key.Y.Cmp(pk1.key.Y)
	if cmp != 0 {
		return false
	}

	return true
}

// Creates a deterministic hash of the public keys
func (pk *PublicKey) Hash() (hash TruncatedHash, err error) {
	if pk.key.X == nil || pk.key.Y == nil {
		return
	}

	combinedKey := append(pk.key.X.Bytes(), pk.key.Y.Bytes()...)
	hash, err = CalculateTruncatedHash(combinedKey)
	return
}

func (pk *PublicKey) GobEncode() (gobPk []byte, err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}

	if pk.key.X == nil || pk.key.Y == nil {
		err = fmt.Errorf("public key not properly initialized")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(pk.key.X)
	if err != nil {
		return
	}
	err = encoder.Encode(pk.key.Y)
	if err != nil {
		return
	}
	gobPk = w.Bytes()
	return
}

func (pk *PublicKey) GobDecode(gobPk []byte) (err error) {
	r := bytes.NewBuffer(gobPk)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&pk.key.X)
	if err != nil {
		return
	}
	err = decoder.Decode(&pk.key.Y)
	if err != nil {
		return
	}
	pk.key.Curve = elliptic.P521() // might there be a way to make this const?
	return
}
