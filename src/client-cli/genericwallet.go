package main

import (
	"client"
	"fmt"
	"state"
)

func pollGenericWallet(c *client.Client, id state.WalletID) {
	var input string
	fmt.Printf("Entering Generic Wallet Mode, for wallet id %x\n", id)
	for {
		fmt.Print("Please enter a wallet action: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHelpWallet()

		case "q", "quit":
			return

		case "d", "download":
			//download(c)

		case "r", "resize":
			resizeGenericWallet(c)

		case "s", "script":
			sendScriptInput(c)

		case "t", "send", "transaction":
			sendFromGenericWallet(c)

		case "u", "upload":
			//uploadToGenericWallet(c)
		}
		input = ""
	}
}
