package main

import (
	"fmt"
	"network"
	"participant"
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

	var directory string
	print("participant directory: ")
	fmt.Scanf("%s", &directory)

	// create a participant
	_, err = participant.CreateParticipant(networkServer, directory)
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
