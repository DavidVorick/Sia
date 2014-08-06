package main

import (
	"client"
	"fmt"
)

func displayServerHelp() {
	fmt.Println(
		"\n",
		"h:\tHelp\n",
		"q:\tReturn to home mode.\n",
		"j:\tCreate a new participant and join an existing quorum.\n",
		"n:\tCreate a new quorum with a bootstrap participant.\n",
	)
}

// serverCreationWalkthrough gets a bunch of input from the user and uses it to
// create a new server.
func serverCreationWalkthrough(c *client.Client) (err error) {
	fmt.Println("No server exists, starting server creation.")

	if !c.IsRouterInitialized() {
		connectWalkthrough(c)
	}

	// Create the server.
	err = c.NewServer()
	if err != nil {
		return
	}

	fmt.Println("Server creation successful!")
	return
}

/*
// joinQuorumWalkthrough gets input about the bootstrap address, the file
// prefix for the particpant, etc. Then the participant is created and
// bootstrapped to an existing quorum.
func joinQuorumWalkthrough(c *client.Client) {
	// ...heh
}
*/

/*
func joinQuorum() {
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
*/

// pollHome is a loop asking users for questions about managing participants.
func pollServer(c *client.Client) {
	var input string
	for {
		fmt.Print("(Server Mode) Please enter a command: ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "help":
			displayServerHelp()

		case "q", "quit", "return":
			return

		case "j", "join":
			fmt.Println("This feature has not been implemented.")
			//joinQuorumWalkthrough(c)

		case "n", "new", "bootstrap":
			fmt.Println("This feature has not been implemented.")
			//newQuorumWalkthrough(s)
		}
	}
}
