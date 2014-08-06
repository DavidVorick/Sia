// Cli Convetions: 'help' is always at the top, followed by quit. All other
// actions are listed in alphabetic order.
package main

import (
	"client"
	"fmt"
)

// The client cli is modal, currently having two states. The first state is the
// 'home' state, where you can request wallets, load wallets, and do general
// high level management of your wallets. The second state is a wallet state,
// which performs actions against a specific wallet.

func printWelcomeMessage() {
	fmt.Println("Sia Client Version 0.0.2")
}

func main() {
	printWelcomeMessage()

	c, err := client.NewClient()
	if err != nil {
		fmt.Println("Error on startup:", err)
	}

	// Check if new client has connected to the network (this would be
	// managed by the config file loader), if not, post some message about
	// not being connected.

	pollHome(c)
}
