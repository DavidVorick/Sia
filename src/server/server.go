package main

import (
	"fmt"
	"network"
	"quorum"
)

func establishQuorum() {
	var port int
	print("Port number: ")
	fmt.Scanf("%d", &port)
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		println(err)
		return
	}
	_, err = quorum.CreateParticipant(networkServer)
	select {}
}

func main() {
	var input string
	for {
		print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			println("unrecognized command")
		case "e":
			establishQuorum()
		case "q":
			return
		}
	}
}
