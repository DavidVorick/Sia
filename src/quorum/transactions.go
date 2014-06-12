package quorum

const (
	SendMaxCost = 6
)

func (q *Quorum) Send(w *wallet, upper uint64, lower uint64, destID WalletID) (cost int) {
	cost += 1
	if w.upperBalance < upper {
		return
	}
	if w.upperBalance == upper && w.lowerBalance < lower {
		return
	}
	cost += 2
	destWallet := q.loadWallet(destID)
	if destWallet == nil {
		return
	}

	cost += 3
	if lower > w.lowerBalance {
		w.upperBalance -= 1
		w.lowerBalance = ^uint64(0) - (lower - w.lowerBalance)
	} else {
		w.lowerBalance -= lower
	}
	w.upperBalance -= upper

	destWallet.upperBalance += upper
	if ^uint64(0)-destWallet.lowerBalance > lower {
		destWallet.lowerBalance += lower
	} else {
		destWallet.upperBalance += 1

		// get lowerBalance to the correct value without ever causing an overflow
		destWallet.lowerBalance = ^uint64(0) - (destWallet.lowerBalance)
		destWallet.lowerBalance += lower
		destWallet.lowerBalance -= ^uint64(0) - (destWallet.lowerBalance)
	}
	return
}
