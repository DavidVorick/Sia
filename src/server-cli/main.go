package main

import (
	"fmt"
	"server"
)

func printWelcomeMessage() {
	fmt.Println("Sia Server Version 0.0.2")
}

// serverCreationWalkthrough gets a bunch of input from the user and uses it to
// create a new server.
func serverCreationWalkthrough() (s *server.Server, err error) {
	fmt.Println("Creating a server...")

	// Get a port number for the RPCServer to listen on.
	var port int
	fmt.Print("What port would you like the server to listen on?")
	_, err = fmt.Scanln("%d", &port)
	if err != nil {
		return
	}

	// Create the server.
	s, err = server.NewServer(port)
	if err != nil {
		return
	}

	fmt.Println("Server creation successful!")
	return
}

func main() {
	printWelcomeMessage()

	s, err := serverCreationWalkthrough()
	if err != nil {
		fmt.Println("Error:", err)
	}

	pollHome(s)
}
