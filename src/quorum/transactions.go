package quorum

func (w *wallet) Send(upper uint64, lower uint64, destination WalletID) {
	// the behavior is undefined if the wallet does not exist.
	// and it's even more curious if the wallet does not exist on another quorum.
	// I guess that if there's enough funding in the send, it just gets bounced.
}
