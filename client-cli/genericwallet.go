package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/state"
)

func downloadGenericWalkthrough(c *client.Client, id state.WalletID) (err error) {
	// Get the name of the filepath to download into.
	var filename string
	fmt.Print("Absolute path to download the file to: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	// yep... that's it.
	err = c.DownloadFile(id, filename)
	if err != nil {
		return
	}

	return
}

// sendCoinGenericWalletWalkthrough walks the user through sending coins from
// their generic wallet.
func sendCoinGenericWalletWalkthrough(c *client.Client, id state.WalletID) (err error) {
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

	err = c.SendCoinGeneric(id, destinationID, state.NewBalance(amount))
	if err != nil {
		return
	}

	return
}

func uploadGenericWalkthrough(c *client.Client, id state.WalletID) (err error) {
	// Get the name of the file to upload.
	var filename string
	fmt.Print("Absolute path of the file to upload: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	// yep... that's it.
	err = c.UploadFile(id, filename)
	if err != nil {
		return
	}

	return
}

// displayGenericWalletHelp prints a list of commands that are available in
// generic wallet mode.
func displayGenericWalletHelp() {
	fmt.Println(
		"h:\tHelp\n",
		"q:\tQuit\n",
		"d:\tDownload the wallet's file.\n",
		"s:\tSend siacoins to another wallet.\n",
		"u:\tUpload a file to the wallet, replacing whatever is currently there.\n",
	)
}

func pollGenericWallet(c *client.Client, id state.WalletID) {
	var input string
	var err error
	for {
		fmt.Println()
		fmt.Printf("(Generic Wallet Mode) Please enter an action for wallet %x: ", id)
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
			err = downloadGenericWalkthrough(c, id)

		case "s", "send", "transaction":
			err = sendCoinGenericWalletWalkthrough(c, id)

		case "u", "upload":
			err = uploadGenericWalkthrough(c, id)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
