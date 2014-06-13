package quorum

const (
	SendMaxCost = 6
)

func (q *Quorum) Send(w *wallet, amount Balance, destID WalletID) (cost int) {
	cost += 1
	if !w.balance.Compare(amount) {
		return
	}
	cost += 2
	destWallet := q.loadWallet(destID)
	if destWallet == nil {
		return
	}

	cost += 3
	w.balance.Subtract(amount)
	destWallet.balance.Add(amount)
	q.saveWallet(destWallet)
	return
}
