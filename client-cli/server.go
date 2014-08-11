package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/state"
)

func displayServerHelp() {
	fmt.Println(
		"\n",
		"h:\tHelp\n",
		"q:\tReturn to home mode.\n",
		"j:\tCreate a new participant and join an existing quorum.\n",
		"n:\tCreate a new quorum with a bootstrap participant.\n",
		"p:\tPrint the status of a server.\n",
	)
}

func serverMetadataWalkthrough(c *client.Client) (err error) {
	fmt.Print("Name of server to fetch status from: ")
	var name string
	_, err = fmt.Scanln(&name)
	if err != nil {
		return
	}

	metadata, err := c.ParticipantMetadata(name)
	if err != nil {
		return
	}

	var siblingString string
	for i := range metadata.Siblings {
		if metadata.Siblings[i].Active {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Active\n", siblingString, i)
		} else {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Inactive\n", siblingString, i)
		}
	}

	fmt.Println(
		"\n",
		name, "status:\n",
		"\tSiblings: \n",
		siblingString,
		"\tHeight:", metadata.Height, "\n",
	)

	return
}

// newQuorumWalkthrough walks the user through creating a new quorum.
func newQuorumWalkthrough(c *client.Client) (err error) {
	fmt.Println("Entering 'New Quorum' walkthorugh")
	fmt.Println("Warning: The client you are using was only intended to work with a single network. This function creates a new Sia network. If you have existing wallets, it's possible that there will be problems.")

	// Get a name for the server, this is what will be used to query the
	// server for status updates in the future.
	var name string
	fmt.Print("Please provide a name for the server: ")
	_, err = fmt.Scanln(&name)
	if err != nil {
		return
	}

	// Get a file prefix for the server. It's possible that one will be
	// specified in the config file, but that's not implemented right now.
	var filepath string
	fmt.Print("Please provide an absolute filepath for the server file directory: ")
	_, err = fmt.Scanln(&filepath)
	if err != nil {
		return
	}

	// Establish a wallet for the first participant in the quorum.
	var sibID state.WalletID
	fmt.Print("Now creating a wallet for the participant to use. Pick an id (hex): ")
	_, err = fmt.Scanln(&sibID)
	if err != nil {
		return
	}

	// Add the wallet as a generic wallet to the client.

	// Create the participant.
	err = c.NewParticipant(name, filepath, sibID)
	if err != nil {
		return
	}

	return
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
			fmt.Println("Warning: this feature is incomplete.")
			err := newQuorumWalkthrough(c)
			if err != nil {
				fmt.Println("Error: ", err)
			}

		case "p", "print", "status":
			err := serverMetadataWalkthrough(c)
			if err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}
}
