package main

import (
	"errors"
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// displayHomeHelp lists all of the options available to the home screen, with
// a description of what they do.
func displayHomeHelp() {
	fmt.Println(
		"\n",
		"h:\tHelp\n",
		"q:\tQuit\n",
		"b:\tBootstrap through an address\n",
		"c:\tConnect to Network\n",
		"l:\tLoad wallet\n",
		"n:\tRequest a new wallet\n",
		"p:\tPrint wallets\n",
		"s:\tSwitch to server mode, creating a server if none yet exists.\n",
		"S:\tSave all wallets\n",
	)
}

// connectWalkthrough requests a port and then calls client.Connect(port),
// initializing the client network router.
func connectWalkthrough(c *client.Client) (err error) {
	// Do nothing if the router is already initialized.
	if c.IsRouterInitialized() {
		err = errors.New("router is already initialized")
		return
	}

	// Get a port.
	var port uint16
	fmt.Print("Port the client should listen on: ")
	_, err = fmt.Scanln(&port)
	if err != nil {
		err = errors.New("invalid port")
		return
	}

	c.Connect(port)
	return
}

// connectWalkthrough guides the user through providing a hostname, port, and
// id which can be used to create a Sia address. Then the connection is
// committed.
func bootstrapToNetworkWalkthrough(c *client.Client) (err error) {
	fmt.Println("Starting connect walkthrough.")
	if !c.IsRouterInitialized() {
		err = connectWalkthrough(c)
		if err != nil {
			return
		}
	}

	fmt.Println("Please indicate the hostname, port, and id that you wish to connect through.")
	var connectAddress network.Address

	// Load the hostname.
	fmt.Print("Hostname: ")
	_, err = fmt.Scanln(&connectAddress.Host)
	if err != nil {
		return
	}

	// Load the port number.
	fmt.Print("Port: ")
	_, err = fmt.Scanln(&connectAddress.Port)
	if err != nil {
		return
	}

	// Load the participant id.
	fmt.Print("ID: ")
	_, err = fmt.Scanln(&connectAddress.ID)
	if err != nil {
		return
	}

	// Call client.Connect using the provided information.
	err = c.BootstrapConnection(connectAddress)
	if err != nil {
		return
	} else {
		fmt.Println("Connection successful.")
	}

	return
}

// loadWallet switches the cli into wallet-mode, where actions are taken
// against a specific wallet.
func loadWallet(c *client.Client) (err error) {
	// Fetch the wallet id from the user.
	var id state.WalletID
	fmt.Print("Wallet ID (hex): ")
	_, err = fmt.Scanln(&id)
	if err != nil {
		return
	}

	// Check that the wallet is available to the client.
	walletType, err := c.WalletType(id)

	// If there was an error, print the error (most likely a not-available
	// error). If the wallet type is recognized, switch to the polling of
	// that wallet type. If the type is not recognized, print an error and
	// return to the home menu.
	if err != nil {
		return
	} else if walletType == "generic" {
		pollGenericWallet(c, id)
	} else {
		err = errors.New("wallet is available, but is of an unknown type.")
		return
	}

	return
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
	fmt.Println()
	fmt.Println("All Stored Wallet IDs:")
	wallets := c.GetWalletIDs()
	for _, id := range wallets {
		fmt.Printf("%x\n", id)
	}
}

// serverModeSwitch will transition the client from being in home mode to being
// in server mode, creating a new server and a new router if necessary.
func serverModeSwitch(c *client.Client) (err error) {
	init := c.IsServerInitialized()
	if !init {
		err = serverCreationWalkthrough(c)
		if err != nil {
			return
		}
	}
	pollServer(c)
	return
}

// pollHome maintains the loop that asks users for actions that are relevant to the home screen.
func pollHome(c *client.Client) {
	var input string
	var err error
	for {
		fmt.Print("(Home) Please enter a command: ")
		_, err = fmt.Scanln(&input)
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

		case "b", "bootstrap":
			err = bootstrapToNetworkWalkthrough(c)

		case "l", "load", "enter":
			err = loadWallet(c)

		case "n", "new", "request":
			fmt.Println("New Wallet is not currently implemented!.")
			// createGenericWallet(c)

		case "p", "ls", "print", "list":
			printWallets(c)

		case "s", "server":
			err = serverModeSwitch(c)

		case "S", "save":
			fmt.Println("Saving all wallets...")
			c.SaveAllWallets()
			fmt.Println("...finished!")
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
