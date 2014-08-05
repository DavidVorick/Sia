package main

import (
	"client"
	"fmt"
)

// displayHomeHelp lists all of the options available to the home screen, with
// a description of what they do.
func displayHomeHelp() {
	fmt.Println(
		"\n",
		"c:\tConnect to Network\n",
		"l:\tLoad wallet\n",
		"n:\tRequest a new wallet\n",
		"p:\tPrint wallets\n",
		"s:\tSave all wallets\n",
		"q:\tQuit\n",
	)
}

// pollHome maintains the loop that asks users for actions that are relevant to the home screen.
func pollHome(c *client.Client) {
	var input string
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHomeHelp()

		case "c", "conncet":
			connectWalkthrough(c)

		case "l", "load", "enter":
			loadWallet(c)

		case "n", "new":
			createGenericWallet(c)

		case "p", "ls", "print", "list":
			listWallets(c)

		case "s", "save":
			c.SaveAllWallets()

		case "q", "quit":
			return
		}
		input = ""
	}
}
