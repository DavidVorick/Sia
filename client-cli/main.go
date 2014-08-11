// Cli Conventions: 'help' is always at the top, followed by quit. All other
// actions are listed in alphabetic order after these two.
//
// The client cli is modal, currently having two states. The first state is the
// 'home' state, where you can request wallets, load wallets, and do general
// high level management of your wallets. The second state is a wallet state,
// which performs actions against a specific wallet. The third state is the
// server state, where you can manage groups of participants that are providing
// storage to the network.
package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
)

// printWelcomeMessage() is the first thing that the user sees when starting
// client-cli, and is the only message that the user is guaranteed to see.
func printWelcomeMessage() {
	fmt.Println("Sia Client Version 0.0.2")
}

// main() prints a welcome message, creates a client, and then shifts into the
// 'home' state.
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
