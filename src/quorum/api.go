package quorum

const (
	CreateWalletMaxCost = 8
	SendMaxCost         = 6
	AddSiblingMaxCost   = 50
)

// CreateWallet takes an id, a Balance, and an initial script and uses
// those to create a new wallet that gets stored in stable memory.
// If a wallet of that id already exists then the process aborts.
func (q *Quorum) CreateWallet(w *Wallet, id WalletID, Balance Balance, initialScript []byte) (cost int) {
	cost += 1
	if !w.Balance.Compare(Balance) {
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
	nw.Balance = Balance
	copy(nw.script, initialScript)
	q.SaveWallet(nw)

	w.Balance.Subtract(Balance)

	return
}

// "Cheat" function for initializing a bootstrap wallet
func (q *Quorum) CreateBootstrapWallet(id WalletID, Balance Balance, initialScript []byte) {
	// check if the new wallet already exists
	wn := q.retrieve(id)
	if wn != nil {
		panic("bootstrap wallet already exists")
	}

	// create a wallet node to insert into the walletTree
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
	nw.Balance = Balance
	copy(nw.script, initialScript)
	q.SaveWallet(nw)
}

func (q *Quorum) Send(w *Wallet, amount Balance, destID WalletID) (cost int) {
	cost += 1
	if !w.Balance.Compare(amount) {
		return
	}
	cost += 2
	destWallet := q.LoadWallet(destID)
	if destWallet == nil {
		return
	}

	cost += 3
	w.Balance.Subtract(amount)
	destWallet.Balance.Add(amount)
	q.SaveWallet(destWallet)
	return
}

// JoinSia is a request that a wallet can submit to make itself a sibling in
// the quorum.
//
// The input is a sibling, a wallet (have to make sure that the wallet used
// as input is the sponsoring wallet...)
//
// Currently, AddSibling tries to add the new sibling to the existing quorum
// and throws the sibling out if there's no space. Once quorums are
// communicating, the AddSibling routine will always succeed.
func (q *Quorum) AddSibling(w *Wallet, s *Sibling) (cost int) {
	cost = 50
	for i := 0; i < QuorumSize; i++ {
		if q.siblings[i] == nil {
			s.index = byte(i)
			s.wallet = w.id
			q.siblings[i] = s
			println("placed hopeful at index", i)
			break
		}
	}
	return
}
