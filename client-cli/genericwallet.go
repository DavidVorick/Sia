package main

/*
import (
	"fmt"

	"github.com/NebulousLabs/Sia/server"
	"github.com/NebulousLabs/Sia/state"
)

func downloadGenericWalletWalkthrough(s *server.Server, gw server.GenericWallet) (err error) {
	// Get the name of the filepath to download into.
	var filename string
	fmt.Print("Absolute path to download the file to: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	err = gw.ID().Download(s, filename)
	if err != nil {
		return
	}

	return
}

// sendCoinGenericWalletWalkthrough walks the user through sending coins from
// their generic wallet.
func sendCoinGenericWalletWalkthrough(s *server.Server, gw server.GenericWallet) (err error) {
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

	err = gw.ID().SendCoin(s, destinationID, state.NewBalance(amount))
	if err != nil {
		return
	}

	return
}

func uploadGenericWalletWalkthrough(s *server.Server, gw server.GenericWallet) (err error) {
	// Get the name of the file to upload.
	var filename string
	fmt.Print("Absolute path of the file to upload: ")
	_, err = fmt.Scanln(&filename)
	if err != nil {
		return
	}

	err = gw.ID().Upload(s, filename)
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

func pollGenericWallet(s *server.Server, gw server.GenericWallet) {
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
			err = downloadGenericWalletWalkthrough(s, gw)

		case "s", "send", "transaction":
			err = sendCoinGenericWalletWalkthrough(s, gw)

		case "u", "upload":
			err = uploadGenericWalletWalkthrough(s, gw)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
*/
