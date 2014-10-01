package main

import (
	//"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

const WalletMenuWidth = 15

// WalletMenuView is a MenuView that lists the wallets available to the user.
type WalletMenuView struct {
	MenuView
}

func newWalletMenuView(parent View) *WalletMenuView {
	wmv := new(WalletMenuView)
	wmv.Parent = parent
	wmv.Title = "Wallets"
	wmv.MenuWidth = WalletMenuWidth
	// load wallet IDs and create views
	wmv.loadWallets()
	return wmv
}

func (wmv *WalletMenuView) Focus() {
	//wmv.loadWallets()
	wmv.MenuView.Focus()
}

func (wmv *WalletMenuView) loadWallets() {
	wids, err := server.GetWallets()
	if err != nil {
		//drawError("Could not load wallets:", err)
		return
	}
	for _, wid := range wids {
		wmv.addWallet(wid)
	}
}

// TODO: to avoid duplicates, merge wallet ids instead of blindly appending
func (wmv *WalletMenuView) addWallet(wid state.WalletID) {
	wmv.Items = append(wmv.Items, wid.String())
	wmv.Windows = append(wmv.Windows, &WalletView{
		DefaultView{Parent: wmv},
	})
}

// A WalletView displays the properties of a Wallet.
type WalletView struct {
	DefaultView
}

func (wv *WalletView) Draw() {

}

func (wv *WalletView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		wv.GiveFocus(wv.Parent)
	}
}
