package main

import (
	"client"
	"fmt"
	"state"
)

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

/*
func download(c *client.Client) {
	var dest string
	fmt.Print("Destination Filepath: ")
	fmt.Scanln(&dest)
	fmt.Println("Downloading File, please wait a few minutes")
	c.Download(c.CurID, dest)
}
*/

/*
func resizeGenericWallet(c *client.Client) {
	var atoms uint16
	var m byte
	fmt.Print("New size (in atoms): ")
	fmt.Scanln(&atoms)
	fmt.Print("Redundancy: ")
	fmt.Scanln(&m)
	err := c.ResizeSector(c.CurID, atoms, m)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Sector resized")
	}
}
*/

/*
func sendScriptInput(c *client.Client) {
	var filename string
	fmt.Print("Input file: ")
	fmt.Scanf("%s", &filename)
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
func sendFromGenericWallet(c *client.Client) {
	var dstID state.WalletID
	var amount uint64
	fmt.Print("Dest Wallet ID (hex): ")
	fmt.Scanf("%x", &dstID)
	fmt.Print("Amount to send (dec): ")
	fmt.Scanln(&amount)
	err := c.SubmitTransaction(c.CurID, dstID, amount)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Transaction successfully submitted")
	}
}
*/

/*
func uploadToGenericWallet(c *client.Client) {
	var filename string
	var k byte
	fmt.Print("Filename: ")
	fmt.Scanln(&filename)
	fmt.Print("K: ")
	fmt.Scanln(&k)
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

func pollGenericWallet(c *client.Client, id state.WalletID) {
	var input string
	fmt.Printf("Entering Generic Wallet Mode")
	for {
		fmt.Print("Please enter an action for wallet %x: ", id)
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "?", "help":
			displayGenericWalletHelp()

		case "q", "quit":
			return

		case "d", "download":
			fmt.Println("Download is not currently implemented.")
			//download(c)

		case "s", "send", "transaction":
			fmt.Println("Sending money is not currently implmeented.")
			//sendFromGenericWallet(c)

		case "u", "upload":
			fmt.Println("Uploading is not currently implemented.")
			//uploadToGenericWallet(c)
		}
		input = ""
	}
}
