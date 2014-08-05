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
		"h:\tHelp\n",
		"q:\tQuit\n",
		"c:\tConnect to Network\n",
		"l:\tLoad wallet\n",
		"n:\tRequest a new wallet\n",
		"p:\tPrint wallets\n",
		"s:\tSave all wallets\n",
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
	_, err := fmt.Scanln("%s", &host)
	if err != nil {
		fmt.Println("Invalid hostname")
		return
	}

	// Load the port number.
	var port int
	fmt.Print("Port: ")
	_, err = fmt.Scanln("%d", &port)
	if err != nil {
		fmt.Println("Invalid port")
		return
	}

	// Load the participant id.
	var id int
	fmt.Print("ID: ")
	_, err = fmt.Scanln("%d", &id)
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
	_, err := fmt.Scanln("%x", &id)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	// Check that the wallet is available to the client.
	walletType, err := c.WalletType(id)

	// If there was an error, print the error (most likely a not-available
	// error). If the wallet type is recognized, switch to the polling of
	// that wallet type. If the type is not recognized, print an error and
	// return to the home menu.
	if err != nil {
		fmt.Println(err)
	} else if walletType == "generic" {
		pollGenericWallet(c, id)
	} else {
		fmt.Println("Wallet is available, but is of an unknown type.")
	}
}

/*
func createGenericWallet(c *client.Client) {
	var id state.WalletID
	var filename string
	fmt.Print("Enter desired Wallet ID (hex): ")
	ERR fmt.Scanln("%x", &id)
	fmt.Print("Script file (blank for default): ")
	ERR fmt.Scanln("%s", &filename)
	var script []byte
	var err error
	if filename != "" {
		script, err = ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}
	err = c.RequestWallet(id, script)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Wallet requested")
	}
}
*/

// printWallets provides a list of every wallet available to the Client.
func printWallets(c *client.Client) {
	fmt.Println("All Stored Wallet IDs:")
	wallets := c.GetWalletIDs()
	for _, id := range wallets {
		fmt.Printf("%x\n", id)
	}
}

// pollHome maintains the loop that asks users for actions that are relevant to the home screen.
func pollHome(c *client.Client) {
	var input string
	for {
		fmt.Print("Please enter a command: ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHomeHelp()

		case "q", "quit":
			return

		case "c", "conncet":
			connectWalkthrough(c)

		case "l", "load", "enter":
			loadWallet(c)

		case "n", "new", "request":
			fmt.Println("New Wallet is not currently implemented!.")
			// createGenericWallet(c)

		case "p", "ls", "print", "list":
			printWallets(c)

		case "s", "save":
			fmt.Println("Saving all wallets...")
			c.SaveAllWallets()
			fmt.Println("...finished!")
		}
		input = ""
	}
}
