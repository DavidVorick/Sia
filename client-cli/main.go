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
