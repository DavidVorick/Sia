package quorum

// chargeWallet subtracts a balance from the wallet depending on the number of
// siblings in the quorum, which is measured by 'multiplier'.
func (q *Quorum) chargeWallets(wn *walletNode, multiplier int) {
	// crawl through the whole tree
	if wn == nil {
		return
	}
	q.chargeWallets(wn.children[0], multiplier)
	q.chargeWallets(wn.children[1], multiplier)

	// load the wallet and deduct storage fees
	w := q.LoadWallet(wn.id)
	weightedPrice := q.storagePrice
	weightedPrice.Multiply(NewBalance(0, uint64(wn.nodeWeight())))
	weightedPrice.Multiply(NewBalance(0, uint64(multiplier)))
	w.Balance.Subtract(weightedPrice)
	q.SaveWallet(w)
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
