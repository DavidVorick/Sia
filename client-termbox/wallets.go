package main

import (
	//"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

const WalletMenuWidth = 15

type WalletMenuView struct {
	MenuView
}

func newWalletMenuView(parent View) View {
	wmv := new(WalletMenuView)
	wmv.Parent = parent
	wmv.Title = "Wallets"
	wmv.MenuWidth = WalletMenuWidth
	// load wallet IDs and create views
	wmv.loadWallets()
	return wmv
}

func (wmv *WalletMenuView) Focus() {
	wmv.hasFocus = true
	wmv.loadWallets()
}

func (wmv *WalletMenuView) loadWallets() {
	wids, err := getWallets()
	if err != nil {
		//drawError("Could not load wallets:", err)
		wmv.Items = []string{"<empty>"}
		wmv.Windows = []View{&WalletView{DefaultView{Parent: wmv}}}
		return
	}
	for _, wid := range wids {
		wmv.addWallet(wid)
	}
}

func (wmv *WalletMenuView) addWallet(wid state.WalletID) {
	wmv.Items = append(wmv.Items, wid.String())
	wmv.Windows = append(wmv.Windows, &WalletView{
		DefaultView{Parent: wmv},
	})
}

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
