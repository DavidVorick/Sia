package delta

import (
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
)

func short(length int) (l, h byte) {
	l = byte(length)
	h = byte(length >> 8)
	return
}

// why isn't this a builtin? Definitely nicer than bytes.Join
func appendAll(slices ...[]byte) []byte {
	var length int
	for _, s := range slices {
		length += len(s)
	}
	all := make([]byte, length)
	i := 0
	for _, s := range slices {
		i += copy(all[i:], s)
	}
	return all
}

// SignScriptInput modifies a ScriptInput to contain a signature of its own
// data. Currently, only the Input and Deadline fields are included in the
// signature.
func SignScriptInput(si *state.ScriptInput, secretKey siacrypto.SecretKey) (err error) {
	sig, err := secretKey.Sign(append(siaencoding.EncUint32(si.Deadline), si.Input...))
	if err != nil {
		return
	}
	si.Input = append(sig[:], si.Input...)
	return
}

// CreateWalletInput returns a script that can be used by the fountain wallet
// to create a new wallet.
func CreateFountainWalletInput(id state.WalletID, s []byte) []byte {
	return append(siaencoding.EncUint64(uint64(id)), s...)
}

// AddSiblingInput returns a signed ScriptInput that calls the AddSibling
// function. It is intended to be passed to a script that transfers execution
// to the input.
func AddSiblingInput(wid state.WalletID, deadline uint32, sib state.Sibling, sk siacrypto.SecretKey) (si state.ScriptInput, err error) {
	si.WalletID = wid
	encSib, err := siaencoding.Marshal(sib)
	if err != nil {
		return
	}
	si.Input = appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to encoded sibling
			0xE4, //       push encoded sibling
			0x41, //       call AddSibling
			0xFF, //       exit
		},
		encSib,
	)
	si.Deadline = deadline

	err = SignScriptInput(&si, sk)
	if err != nil {
		return
	}
	return
}

// TransactionInput returns a script that calls the Send function. It is
// intended to be passed to a script that transfers execution to the input.
func SendCoinInput(dest state.WalletID, amount state.Balance) []byte {
	return appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to dest
			0x34, 0x08, // push dest
			0x34, 0x10, // push balance
			0x43, //       call Send
			0xFF, //       exit
		},
		siaencoding.EncUint64(uint64(dest)),
		amount[:],
	)
}

// UpdateSectorInput returns a script that calls the UpdateSector function. It
// is intended to be passed to a script that transfers execution to the input.
func UpdateSectorInput(su state.SectorUpdate) []byte {
	hlen := byte(siacrypto.HashSize)
	slen := state.QuorumSize * hlen // TODO: this must be fixed for QuorumSize > 8
	hashset := make([]byte, slen)
	for i, h := range su.HashSet {
		copy(hashset[i*siacrypto.HashSize:], h[:])
	}
	return appendAll(
		[]byte{
			0xE6, 0xFF, // move data pointer to parent hash
			0x34, hlen, // push parent hash
			0x34, 0x02, // push atoms
			0x34, 0x01, // push k
			0x34, 0x01, // push d
			0x34, slen, // push hashset
			0x34, 0x01, // push confreq
			0x34, 0x04, // push deadline
			0x44, //       call UpdateSector
			0xFF, //       exit
		},
		su.ParentHash[:],
		siaencoding.EncUint16(su.Atoms),
		[]byte{su.K, su.D},
		hashset,
		[]byte{su.ConfirmationsRequired},
		siaencoding.EncUint32(su.Deadline),
	)
}

// The FountainScript acts as a fountain, and can be called to spawn a new
// wallet with a minimal starting balance that will the economy to get flowing.
var FountainScript = []byte{
	0x34, 0x08, //       00 push 8 bytes of input (wallet id)
	0x02, 0xA8, 0x61, // 02 push 25000 (balance)
	0xE4, //             05 push script
	0x42, //             06 call CreateWallet
	0xFF, //             07 exit
}

// DefaultScript returns a script that verifies a signature, and transfers
// control to the input if the verification was successful.
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	keyl, sigl := byte(siacrypto.PublicKeySize), byte(siacrypto.SignatureSize)
	negl, negh := short(-siacrypto.PublicKeySize)
	return append([]byte{
		0x33, negl, negh, // 00 move data pointer to public key
		0x34, keyl, //       03 push public key
		0x34, sigl, //       04 push signature
		0x46, //             06 push deadline
		0xE4, //             07 push input
		0x23, //             08 concatenate deadline and input
		0x40, //             09 verify signature
		0xE5, //             10 if invalid signature, reject
		0x38, //             11 execute input
	}, publicKey[:]...)
}
