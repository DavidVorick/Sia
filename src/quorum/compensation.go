package quorum

// chargeWallets subtracts a balance from the wallet depending on the number of
// siblings in the quorum, which is measured by 'multiplier'.
func (s *State) chargeWallets(wn *walletNode, multiplier int) {
	// Charge every wallet in the wallet tree.
	if wn == nil {
		return
	}
	s.chargeWallets(wn.children[0], multiplier)
	s.chargeWallets(wn.children[1], multiplier)

	// Load the wallet and calculate the weighted price, which is the cost of
	// storing the atoms on all of the siblings currently active in the quorum.
	w, err := s.LoadWallet(wn.id)
	if err != nil {
		panic(err)
	}
	weightedPrice := s.Metadata.StoragePrice
	weightedPrice.Multiply(NewBalance(0, uint64(wn.nodeWeight())))
	weightedPrice.Multiply(NewBalance(0, uint64(multiplier)))

	// If the wallet does not have enough money to pay for the storage it
	// consumes between this block and next block, the wallet is deleted.
	if weightedPrice.Compare(w.Balance) == 1 {
		// Wallet has run out of funds, purge from the network.
	} else {
		w.Balance.Subtract(weightedPrice)
		s.SaveWallet(w)
	}
}

// ExecuteCompensation() is called between each block. Money is deducted from
// wallets according to how much storage they are using, and money is added to
// siblings according to how much storage is in use.
func (s *State) ExecuteCompensation() {
	if s.walletRoot == nil {
		return
	}

	// Determine how much to pay each sibling, which is the weight of the quorum
	// multiplied by the storage price.
	compensation := s.Metadata.StoragePrice
	compensation.Multiply(NewBalance(0, uint64(s.walletRoot.weight)))

	// Pay each sibling the appropriate compensation.
	var siblings int
	for i := range s.Metadata.Siblings {
		if s.Metadata.Siblings[i] == nil {
			continue
		}

		w, err := s.LoadWallet(s.Metadata.Siblings[i].wallet)
		if err != nil {
			panic(err)
		}
		w.Balance.Add(compensation)
		s.SaveWallet(w)
		siblings++
	}

	// Call a helper function to charge all the wallets for the storage they have
	// consumed.
	s.chargeWallets(s.walletRoot, siblings)
}
