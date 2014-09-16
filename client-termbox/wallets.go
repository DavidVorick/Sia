package main

import (
	"math/rand"

	//"github.com/NebulousLabs/Sia/network"
	//"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

type WalletsView struct {
	Parent   View
	hasFocus bool
}

func (wv *WalletsView) Draw(r Rectangle) {
	for y := r.MinY; y < r.MaxY; y++ {
		for x := r.MinX; x < r.MaxX; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+1)
		}
	}
}

func (wv *WalletsView) HandleKey(key termbox.Key) {

}

func (wv *WalletsView) GiveFocus() {
	wv.hasFocus = true
}

/*
func drawWallets(startColumn int) {
		// Fetch a list of wallets from the server.
		var ids *[]state.WalletID
		err := networkState.Router.SendMessage(network.Message{
			Dest: networkState.ServerAddress,
			Proc: "Server.WalletIDs",
			Args: struct{}{},
			Resp: &ids,
		})
		if err != nil {
			// Draw an error message.
			// return
		}
}
*/
