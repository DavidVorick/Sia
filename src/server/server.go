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

func printHelp() {
	println()
	println("h - help")
	println("help - help")
	println("e - establish a participant, who will either create or join a quorum depending on the bootstrap settings")
	println("q - quit")
	println()
}

func main() {
	var input string
	println("Sia Server Version 0.0.0.2")
	for {
		// grab some input
		print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			println("unrecognized command")

		// e: create a participant and bootstrap to a quorum
		case "e":
			establishQuorum()

		// q: quit the program
		case "q":
			return

		case "h":
			printHelp()

		case "help":
			printHelp()
		}
	}
	println()
}
