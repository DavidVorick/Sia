package state

// chargeWallets subtracts a balance from the wallet depending on the number of
// siblings in the quorum, which is measured by 'multiplier'.
func (s *State) chargeWallets(wn *walletNode, multiplier int) (quorumWeight uint64) {
	// Charge every wallet in the wallet tree.
	if wn == nil {
		return
	}
	quorumWeight += s.chargeWallets(wn.children[0], multiplier)
	quorumWeight += s.chargeWallets(wn.children[1], multiplier)

	// Load the wallet and calculate the weighted price, which is the cost of
	// storing the atoms on all of the siblings currently active in the quorum.
	w, err := s.LoadWallet(wn.id)
	if err != nil {
		panic(err)
	}
	weightedPrice := s.Metadata.StoragePrice
	weightedPrice.Multiply(NewBalance(0, uint64(w.CompensationWeight())))
	weightedPrice.Multiply(NewBalance(0, uint64(multiplier)))

	// If the wallet does not have enough money to pay for the storage it
	// consumes between this block and next block, the wallet is deleted.
	if weightedPrice.Compare(w.Balance) == 1 {
		s.RemoveWallet(w.ID)
	} else {
		w.Balance.Subtract(weightedPrice)
		quorumWeight += uint64(w.CompensationWeight())
		s.SaveWallet(w)
	}
	return
}

// ExecuteCompensation is called between each block. Money is deducted from
// wallets according to how much storage they are using, and money is added to
// siblings according to how much storage is in use.
func (s *State) ExecuteCompensation() {
	if s.walletRoot == nil {
		return
	}

	// Count the number of siblings receiving compensation.
	var siblings int
	for i := range s.Metadata.Siblings {
		if s.Metadata.Siblings[i].Active() {
			siblings++
		}
	}

	// Call a helper function to charge all the wallets for the storage they have
	// consumed. chargeWallets must be called before the siblings are
	// compensated, so that the siblins don't get compensated for wallets that
	// have been deleted by chargeWallets.
	quorumWeight := s.chargeWallets(s.walletRoot, siblings)

	// Compensate each sibling.
	compensation := s.Metadata.StoragePrice
	compensation.Multiply(NewBalance(0, quorumWeight))
	for i := range s.Metadata.Siblings {
		if !s.Metadata.Siblings[i].Active() {
			continue
		}

		w, err := s.LoadWallet(s.Metadata.Siblings[i].WalletID)
		if err != nil {
			panic(err)
		}
		w.Balance.Add(compensation)
		s.SaveWallet(w)
	}
}
