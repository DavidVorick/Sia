package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/server"
	"github.com/NebulousLabs/Sia/state"
)

func downloadGenericWalletWalkthrough(c *server.Client, gw server.GenericWallet) (err error) {
	// Get the name of the filepath to download into.
	var filename string
	fmt.Print("Absolute path to download the file to: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	err = gw.ID().Download(c, filename)
	if err != nil {
		return
	}

	return
}

// sendCoinGenericWalletWalkthrough walks the user through sending coins from
// their generic wallet.
func sendCoinGenericWalletWalkthrough(c *server.Client, gw server.GenericWallet) (err error) {
	// Get a destination and an amount
	var destinationID state.WalletID
	var amount uint64
	fmt.Print("Destination Wallet ID: ")
	_, err = fmt.Scanln(&destinationID)
	if err != nil {
		return
	}
	fmt.Print("Amount to send: ")
	_, err = fmt.Scanln(&amount)
	if err != nil {
		return
	}

	err = gw.ID().SendCoin(c, destinationID, state.NewBalance(amount))
	if err != nil {
		return
	}

	return
}

func uploadGenericWalletWalkthrough(c *server.Client, gw server.GenericWallet) (err error) {
	// Get the name of the file to upload.
	var filename string
	fmt.Print("Absolute path of the file to upload: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	err = gw.ID().Upload(c, filename)
	if err != nil {
		return
	}

	return
}

// displayGenericWalletHelp prints a list of commands that are available in
// generic wallet mode.
func displayGenericWalletHelp() {
	fmt.Println(
		" h:\tHelp\n",
		"q:\tQuit\n",
		"d:\tDownload the wallet's file.\n",
		"s:\tSend siacoins to another wallet.\n",
		"u:\tUpload a file to the wallet, replacing whatever is currently there.\n",
	)
}

func pollGenericWallet(c *server.Client, gw server.GenericWallet) {
	var input string
	var err error
	for {
		fmt.Println()
		fmt.Printf("(Generic Wallet Mode) Please enter an action for wallet %x: ", gw.ID())
		_, err = fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
		fmt.Println()

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "?", "help":
			displayGenericWalletHelp()

		case "q", "quit":
			return

		case "d", "download":
			err = downloadGenericWalletWalkthrough(c, gw)

		case "s", "send", "transaction":
			err = sendCoinGenericWalletWalkthrough(c, gw)

		case "u", "upload":
			err = uploadGenericWalletWalkthrough(c, gw)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
