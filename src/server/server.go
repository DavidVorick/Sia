package main

import (
	"fmt"
	"network"
	"participant"
)

func joinQuorum() {
	// read port number
	var port int
	fmt.Print("Port to listen on: ")
	fmt.Scanf("%d", &port)

	// create a message router
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer networkServer.Close()

	// read and set bootstrap address
	var hostname string
	fmt.Print("Bootstrap hostname: ")
	fmt.Scanf("%s", &hostname)
	fmt.Print("Bootstrap port: ")
	fmt.Scanf("%d", &port)
	participant.BootstrapAddress.Host = hostname
	participant.BootstrapAddress.Port = port
	err = networkServer.Ping(&participant.BootstrapAddress)
	if err != nil {
		fmt.Println("Failed to ping bootstrap:", err)
		return
	}

	var directory string
	fmt.Print("participant directory: ")
	fmt.Scanf("%s", &directory)

	// create a participant
	_, err = participant.CreateParticipant(networkServer, directory)
	if err != nil {
		fmt.Println(err)
		return
	}
	select {}
}

func establishQuorum() {
	// read and set port number
	var port int
	fmt.Print("Port to listen on: ")
	fmt.Scanf("%d", &port)
	participant.BootstrapAddress.Port = port

	// create a message router
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer networkServer.Close()

	var directory string
	fmt.Print("participant directory: ")
	fmt.Scanf("%s", &directory)

	// create a participant
	_, err = participant.CreateParticipant(networkServer, directory)
	if err != nil {
		fmt.Println(err)
		return
	}
	select {}
}

func printHelp() {
	fmt.Println(`
h - help
help - help
j - join an existing quorum
e - establish a new quorum
q - quit
`)
}

func main() {
	var input string
	fmt.Println("Sia Server Version 0.0.0.2")
	for {
		// grab some input
		print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		// j: create a participant and bootstrap to a quorum
		case "j":
			joinQuorum()

		// e: create a participant and bootstrap to a quorum
		case "e":
			establishQuorum()

		// q: quit the program
		case "q":
			return

		case "h", "help":
			printHelp()
		}
	}
}
