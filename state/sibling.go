package state

const (
	// SiblingPassiveWindow is the number of blocks that a sibling is
	// allowed to be passive.
	SiblingPassiveWindow = 2
)
// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings. The Status of a sibling
// indicates it's standing with the quorum. ^byte(0) indicates that the sibling
// is 'Inactive', and that there are no hosts filling that position. A standing
// of '5-1' indicates that the sibling is 'Passive', with the number indicating
// how many compiles until the sibling becomes active. A passive sibling is
// sent updates during consensus, and its signatures are accepted during
// consensus, but it's heartbeats are not included into block compilation.
// Passive sibling will not be included in compensation. An active sibling is a
// full sibing that _must_ participate in consensus and provide updates to the
// network.
type Sibling struct {
	Status    byte
	Index     byte
	Address   network.Address
	PublicKey siacrypto.PublicKey
	WalletID  WalletID
}

// Active returns true if the sibling is a fully active member of the quorum
// according to the status variable, false if the sibling is passive or
// inactive.
func (sib Sibling) Active() bool {
	return sib.Status == 0
}

// Inactive returns true if the sibling is inactive, and retuns false if the
// sibling is active or passive.
func (sib Sibling) Inactive() bool {
	return sib.Status == ^byte(0)
}
