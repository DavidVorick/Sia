package quorum

const (
	CreateWalletMaxCost = 8
	SendMaxCost         = 6
)

// CreateWallet takes an id, a balance, a number of script atom, and an initial
// script and uses those to create a new wallet that gets stored in stable
// memory. If a wallet of that id already exists then the process aborts.
func (q *Quorum) CreateWallet(w *Wallet, id WalletID, balance Balance, initialScript []byte) (cost int) {
	cost += 1
	if !w.balance.Compare(balance) {
		return
	}

	// check if the new wallet already exists
	cost += 2
	wn := q.retrieve(id)
	if wn != nil {
		return
	}

	// create a wallet node to insert into the walletTree
	cost += 5
	wn = new(walletNode)
	wn.id = id
	wn.weight = 1
	tmp := len(initialScript)
	tmp -= 1024
	for tmp > 0 {
		wn.weight += 1
		tmp -= 4096
	}
	q.insert(wn)

	// fill out a basic wallet struct from the inputs
	nw := new(Wallet)
	nw.id = id
	nw.balance = balance
	copy(nw.script, initialScript)
	q.SaveWallet(nw)

	w.balance.Subtract(balance)

	return
}

func (q *Quorum) Send(w *Wallet, amount Balance, destID WalletID) (cost int) {
	cost += 1
	if !w.balance.Compare(amount) {
		return
	}
	cost += 2
	destWallet := q.LoadWallet(destID)
	if destWallet == nil {
		return
	}

	cost += 3
	w.balance.Subtract(amount)
	destWallet.balance.Add(amount)
	q.SaveWallet(destWallet)
	return
}
