package main

import (
	//"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

const WalletMenuWidth = 15

type WalletsMenuView struct {
	MenuWindow
}

func newWalletMenuView(parent View) View {
	wmv := &WalletsMenuView{
		MenuWindow: MenuWindow{Parent: parent,
			Title:     "Wallets",
			MenuWidth: WalletMenuWidth,
			sel:       0,
			hasFocus:  false,
		},
	}
	// load wallet IDs and create views
	wmv.loadWallets()
	return wmv
}

func (wmv *WalletsMenuView) GiveFocus() {
	wmv.hasFocus = true
	wmv.loadWallets()
}

func (wmv *WalletsMenuView) loadWallets() {
	wids, err := getWallets()
	if err != nil {
		//drawError("Could not load wallets:", err)
		wmv.Items = []string{"<empty>"}
		wmv.Windows = []View{&WalletView{Parent: wmv}}
		return
	}
	for _, wid := range wids {
		wmv.addWallet(wid)
	}
}

func (wmv *WalletsMenuView) addWallet(wid state.WalletID) {
	wmv.Items = append(wmv.Items, wid.String())
	wmv.Windows = append(wmv.Windows, &WalletView{
		Parent:   wmv,
		hasFocus: false,
	})
}

type WalletView struct {
	Rectangle
	Parent   View
	hasFocus bool
}

func (wv *WalletView) SetDims(r Rectangle) {
	wv.Rectangle = r
}

func (wv *WalletView) GiveFocus() {
	wv.hasFocus = true
}

func (wv *WalletView) Draw() {

}

func (wv *WalletView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		wv.hasFocus = false
		wv.Parent.GiveFocus()
	}
}
