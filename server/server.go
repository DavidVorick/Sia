package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

func joinQuorum() {
	// read port number
	var port uint16
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
	var bootstrap network.Address
	fmt.Print("Bootstrap hostname: ")
	fmt.Scanf("%s", &bootstrap.Host)
	fmt.Print("Bootstrap port: ")
	fmt.Scanf("%d", &bootstrap.Port)
	fmt.Print("Bootstrap ID: ")
	fmt.Scanf("%d", &bootstrap.ID)
	err = networkServer.Ping(bootstrap)
	if err != nil {
		fmt.Println("Failed to ping bootstrap:", err)
		return
	}

	// read wallet ID
	var sibID uint64
	fmt.Print("Wallet ID: ")
	fmt.Scanf("%d", &sibID)

	// obtain a wallet
	err = networkServer.SendMessage(network.Message{
		Dest: bootstrap,
		Proc: "Participant.AddScriptInput",
		Args: delta.CreateWalletInput(sibID, []byte{0x2F}), // need better default script
		Resp: nil,
	})
	if err != nil {
		fmt.Println("Failed to obtain wallet:", err)
		return
	}

	var directory string
	fmt.Print("participant directory: ")
	fmt.Scanf("%s", &directory)

	// create a participant
	_, err = consensus.CreateJoiningParticipant(networkServer, directory, state.WalletID(sibID), []network.Address{bootstrap})
	if err != nil {
		fmt.Println(err)
		return
	}

	// block forever
	select {}
}

func establishQuorum() {
	// read and set port number
	var port uint16
	fmt.Print("Port to listen on: ")
	fmt.Scanf("%d", &port)

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

	var sibID state.WalletID
	fmt.Print("Wallet ID: ")
	fmt.Scanf("%d", &sibID)

	// create a participant
	_, err = consensus.CreateBootstrapParticipant(networkServer, directory, sibID)
	if err != nil {
		fmt.Println(err)
		return
	}
	select {}
}

func printHelp() {
	fmt.Println(`
h - help
j - join an existing quorum
e - establish a new quorum
q - quit
`)
}

func main() {
	var input string
	fmt.Println("Sia Server Version 0.0.1")
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
