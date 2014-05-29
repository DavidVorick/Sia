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

type PublicKey ecdsa.PublicKey
type SecretKey ecdsa.PrivateKey
type Signature struct {
	R, S *big.Int
}
type Hash [HashSize]byte
type TruncatedHash [TruncatedHashSize]byte

// Compare returns true if the keys are composed of the same integer values
// Compare returns false if any sub-value is nil
func (pk0 *PublicKey) Compare(pk1 *PublicKey) bool {
	epk0 := (*ecdsa.PublicKey)(pk0)
	epk1 := (*ecdsa.PublicKey)(pk1)

	// return false if either value is nil
	if epk0 == nil || epk1 == nil {
		return false
	}

	// return false if either sub-value is nil
	if epk0.X == nil || epk0.Y == nil || epk1.X == nil || epk1.Y == nil {
		return false
	}

	cmp := epk0.X.Cmp(epk1.X)
	if cmp != 0 {
		return false
	}

	cmp = epk0.Y.Cmp(epk1.Y)
	if cmp != 0 {
		return false
	}

	return true
}

func (pk *PublicKey) Hash() (hash TruncatedHash, err error) {
	epk := (*ecdsa.PublicKey)(pk)
	if epk.X == nil || epk.Y == nil {
		return
	}

	combinedKey := append(epk.X.Bytes(), epk.Y.Bytes()...)
	hash, err = CalculateTruncatedHash(combinedKey)
	return
}

func (pk *PublicKey) GobEncode() (gobPk []byte, err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}

	epk := (*ecdsa.PublicKey)(pk)
	if epk.X == nil || epk.Y == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(epk.X)
	if err != nil {
		return
	}
	err = encoder.Encode(epk.Y)
	if err != nil {
		return
	}
	gobPk = w.Bytes()
	return
}

func (pk *PublicKey) GobDecode(gobPk []byte) (err error) {
	epk := (*ecdsa.PublicKey)(pk)
	r := bytes.NewBuffer(gobPk)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&epk.X)
	if err != nil {
		return
	}
	err = decoder.Decode(&epk.Y)
	if err != nil {
		return
	}
	epk.Curve = elliptic.P521() // might there be a way to make this const?
	return
}
