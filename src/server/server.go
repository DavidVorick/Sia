package main

import (
	"fmt"
	"network"
	"quorum"
)

func establishQuorum() {
	// grab a port number
	var port int
	print("Port number: ")
	fmt.Scanf("%d", &port)

	// create a message router
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		fmt.Println(err)
		return
	}

	// create a participant
	_, err = quorum.CreateParticipant(networkServer)
	if err != nil {
		fmt.Println(err)
	}
	select {}
}

func main() {
	var input string
	for {
		// grab some input
		print("Please enter a command: ")
		fmt.Scanln(&input)

		// switch on the input as a command
		switch input {

		// get more input if input is invalid
		default:
			println("unrecognized command")

		// e: create a participant and bootstrap to a quorum
		case "e":
			establishQuorum()

		// q: quit the program
		case "q":
			return
		}
	}
}
