package main

import (
	"client"
	"fmt"
	"state"
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

// connectWalkthrough guides the user through providing a hostname, port, and
// id which can be used to create a Sia address. Then the connection is
// committed.
func connectWalkthrough(c *client.Client) {
	// Eliminate all existing connections.
	c.Disconnect()

	fmt.Println("Please indicate the hostname, port, and id that you wish to connect to.")

	// Load the hostname.
	var host string
	fmt.Print("Hostname: ")
	_, err := fmt.Scanf("%s", &host)
	if err != nil {
		fmt.Println("Invalid hostname")
		return
	}

	// Load the port number.
	var port int
	fmt.Print("Port: ")
	_, err = fmt.Scanf("%d", &port)
	if err != nil {
		fmt.Println("Invalid port")
		return
	}

	// Load the participant id.
	var id int
	fmt.Print("ID: ")
	_, err = fmt.Scanf("%d", &id)
	if err != nil {
		fmt.Println("Invalid id")
		return
	}

	// Call client.Connect using the provided information.
	err = c.Connect(host, port, id)
	if err != nil {
		fmt.Println("Error while connecting:", err)
	} else {
		fmt.Println("Connection successful.")
	}
}

// loadWallet switches the cli into wallet-mode, where actions are taken
// against a specific wallet.
func loadWallet(c *client.Client) {
	// Fetch the wallet id from the user.
	var id state.WalletID
	fmt.Print("Wallet ID (hex): ")
	fmt.Scanf("%x", &id)

	// Check that the wallet is available to the client.
	walletType, err := c.WalletType(id)

	// If the wallet is available, switch to the mode associated with that
	// wallet type. Otherwise, print the error.
	if err != nil {
		pollGenericWallet(c)
	} else {
		fmt.Println(err)
	}
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
