package main

import (
	//"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

const WalletMenuWidth = 15

// WalletMenuMVC is a MenuMVC that lists the wallets available to the user.
type WalletMenuMVC struct {
	MenuMVC
}

func newWalletMenuMVC(parent MVC) *WalletMenuMVC {
	wm := new(WalletMenuMVC)
	wm.Parent = parent
	wm.Title = "Wallets"
	wm.MenuWidth = WalletMenuWidth
	// load wallet IDs and create views
	wm.loadWallets()
	return wm
}

func (wm *WalletMenuMVC) Focus() {
	wm.loadWallets()
	wm.MenuMVC.Focus()
}

func (wm *WalletMenuMVC) loadWallets() {
	wids, err := server.WalletIDs()
	if err != nil {
		drawError("Could not load wallets:", err)
		return
	}

	// clear existing wallets
	// (see comment on loadParticipants)
	wm.Items = []string{}
	wm.Windows = []MVC{}

	for _, wid := range wids {
		wm.addWallet(wid)
	}
}

func (wm *WalletMenuMVC) addWallet(wid state.WalletID) {
	wm.Items = append(wm.Items, wid.String())
	wm.Windows = append(wm.Windows, &WalletMVC{
		DefaultMVC{Parent: wm},
	})
}

// A WalletMVC displays the properties of a Wallet.
type WalletMVC struct {
	DefaultMVC
}

func (wv *WalletMVC) Draw() {

}

func (wv *WalletMVC) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		wv.GiveFocus(wv.Parent)
	}
}
