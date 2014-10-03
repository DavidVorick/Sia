package main

import (
	"fmt"

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
	// set dimensions of children
	wm.SetDims(wm.Rectangle)
}

func (wm *WalletMenuMVC) addWallet(wid state.WalletID) {
	w := new(WalletMVC)
	w.Parent = wm
	if err := server.Wallet(wid, &w.wallet); err != nil {
		drawError(fmt.Sprintf("Could not fetch wallet %v: %v", wid, err))
		return
	}
	wm.Items = append(wm.Items, wid.String())
	wm.Windows = append(wm.Windows, w)
}

// A WalletMVC displays the properties of a Wallet.
type WalletMVC struct {
	DefaultMVC
	id     state.WalletID
	wallet state.Wallet
}

func (wv *WalletMVC) Draw() {
	drawString(wv.MinX+1, wv.MinY+1, fmt.Sprint("Balance: ", wv.wallet.Balance))
	drawString(wv.MinX+1, wv.MinY+2, fmt.Sprint("Atoms: ", wv.wallet.Sector.Atoms))
	drawString(wv.MinX+1, wv.MinY+3, fmt.Sprint("Hash: ", wv.wallet.Sector.Hash()))
}

func (wv *WalletMVC) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		wv.GiveFocus(wv.Parent)
	}
}
