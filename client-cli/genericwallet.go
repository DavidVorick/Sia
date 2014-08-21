package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/state"
)

/*
func downloadWalkthrough(c *client.Client) {
	var dest string
	fmt.Print("Destination Filepath: ")
	ERR fmt.Scanln(&dest)
	fmt.Println("Downloading File, please wait a few minutes")
	c.Download(c.CurID, dest)
}
*/

/*
func resizeGenericWalletWalkthrough(c *client.Client) {
	var atoms uint16
	var m byte
	fmt.Print("New size (in atoms): ")
	ERR fmt.Scanln(&atoms)
	fmt.Print("Redundancy: ")
	ERR fmt.Scanln(&m)
	err := c.ResizeSector(c.CurID, atoms, m)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Sector resized")
	}
}
*/

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

	balance := state.NewBalance(amount)
	err = c.SendCoinGeneric(id, destinationID, balance)
	if err != nil {
		return
	}

	return
}

/*
func sendScriptInputWalkthrough(c *client.Client) {
	var filename string
	fmt.Print("Input file: ")
	ERR fmt.Scanln("%s", &filename)
	// read script from file
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = c.SendCustomInput(c.CurID, input)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Script input sent")
	}
}
*/

/*
func uploadToGenericWalletWalkthrough(c *client.Client) {
	var filename string
	var k byte
	fmt.Print("Filename: ")
	ERR fmt.Scanln(&filename)
	fmt.Print("K: ")
	ERR fmt.Scanln(&k)
	atomsRequired, err := client.CalculateAtoms(filename, k)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Atoms Required: %v\n", atomsRequired)
	}
	fmt.Println("Attempting to Upload File, please wait a few minutes (longer for large files).")
	c.UploadFile(c.CurID, filename, k)
}
*/

// displayGenericWalletHelp prints a list of commands that are available in
// generic wallet mode.
func displayGenericWalletHelp() {
	fmt.Println(
		"\n",
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
		fmt.Printf("(Generic Wallet Mode) Please enter an action for wallet %x: ", id)
		_, err = fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "?", "help":
			displayGenericWalletHelp()

		case "q", "quit":
			return

		case "d", "download":
			fmt.Println("Download is not currently implemented.")
			//download(c, id)

		case "s", "send", "transaction":
			sendCoinGenericWalletWalkthrough(c, id)

		case "u", "upload":
			fmt.Println("Uploading is not currently implemented.")
			//uploadToGenericWallet(c, id)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
