package quorum

// chargeWallet subtracts a balance from the wallet depending on the number of
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
	w := s.LoadWallet(wn.id)
	weightedPrice := s.storagePrice
	weightedPrice.Multiply(NewBalance(0, uint64(wn.nodeWeight())))
	weightedPrice.Multiply(NewBalance(0, uint64(multiplier)))

	// If the wallet does not have enough money to pay for the storage it
	// consumes between this block and next block, the wallet is deleted.
	if weighted.Compare(w.Balace) == 1 {
		// Wallet has run out of funds, purge from the network.
	} else {
		w.Balance.Subtract(weightedPrice)
		s.SaveWallet(w)
	}
}

// ExecuteCompensation() is called between each block. Money is deducted from wallets
func (q *Quorum) ExecuteCompensation() {
	if q.walletRoot == nil {
		return
	}

	compensation := q.storagePrice
	compensation.Multiply(NewBalance(0, uint64(q.walletRoot.weight)))
	var siblings int
	for i := range q.siblings {
		if q.siblings[i] == nil {
			continue
		}

		w := q.LoadWallet(q.siblings[i].wallet)
		w.Balance.Add(compensation)
		q.SaveWallet(w)
		siblings++
	}

	q.chargeWallets(q.walletRoot, siblings)
}
