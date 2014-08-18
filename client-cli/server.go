package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/state"
)

func printMetadata(metadata state.Metadata) {
	var siblingString string
	for i := range metadata.Siblings {
		if metadata.Siblings[i].Active() {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Active\n", siblingString, i)
		} else if metadata.Siblings[i].Inactive() {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Inactive\n", siblingString, i)
		} else {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Passive for %v more compiles.", siblingString, i, metadata.Siblings[i].Status)
		}
	}

	fmt.Println(
		"\tSiblings: \n",
		siblingString,
		"\tHeight:", metadata.Height, "\n",
		"\tRecent Snapshot:", metadata.RecentSnapshot, "\n",
	)
}

// serverNameAndFilepathWalkthrough is a helper function that gets and returns
// the name and folderpath of a server.
func serverNameAndFilepathWalkthrough(c *client.Client) (name string, filepath string, err error) {
	fmt.Println()

	// Get a name for the server, this is what will be used to query the
	// server for status updates in the future.
	fmt.Print("Please provide a name for the server: ")
	_, err = fmt.Scanln(&name)
	if err != nil {
		return
	}

	// Get a file prefix for the server. It's possible that one will be
	// specified in the config file, but that's not implemented right now.
	fmt.Print("Please provide an absolute filepath for the server file directory: ")
	_, err = fmt.Scanln(&filepath)
	if err != nil {
		return
	}

	return
}

// joinQuorumWalkthrough gets input about the bootstrap address, the file
// prefix for the particpant, etc. Then the participant is created and
// bootstrapped to an existing quorum.
func joinQuorumWalkthrough(c *client.Client) (err error) {
	fmt.Println()
	fmt.Println("Entering 'Join Quorum' Walkthrough")

	// Get a name and filepath.
	name, filepath, err := serverNameAndFilepathWalkthrough(c)
	if err != nil {
		return
	}

	// Get a tether id for the participant.
	// Establish a wallet for the first participant in the quorum.
	var tetherID state.WalletID
	fmt.Print("Which wallet would you like to use as the tether wallet (must be a generic wallet): ")
	_, err = fmt.Scanln(&tetherID)
	if err != nil {
		return
	}

	fmt.Println("Creating a joining participant and bootstrapping it to the network. Note that this will take several blocks.")
	err = c.NewJoiningParticipant(name, filepath, tetherID)
	if err != nil {
		return
	}
	fmt.Println("Participant creation and quorum joining successful.")

	return
}

// newQuorumWalkthrough walks the user through creating a new quorum.
func newQuorumWalkthrough(c *client.Client) (err error) {
	fmt.Println()
	fmt.Println("Entering 'New Quorum' walkthorugh")
	fmt.Println("Warning: The client you are using was only intended to work with a single network. This function creates a new Sia network. If you have existing wallets, it's possible that there will be problems.")
	fmt.Println("Double Warning: This feature is really only be meant to be used by developers. A lot can go wrong, just please be careful and realized that you were warned if bad stuff happens.")

	name, filepath, err := serverNameAndFilepathWalkthrough(c)
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

	// Create the participant.
	err = c.NewBootstrapParticipant(name, filepath, sibID)
	if err != nil {
		return
	}

	return
}

// serverCreationWalkthrough gets a bunch of input from the user and uses it to
// create a new server.
func serverCreationWalkthrough(c *client.Client) (err error) {
	fmt.Println()
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

// serverMetadataWalkthrough collects the information needed to pull up the
// metadata of a participant (IE the name), then collects the metadata, formats
// it into human-readable form, and prints it.
func serverMetadataWalkthrough(c *client.Client) (err error) {
	fmt.Println()
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

	fmt.Println(
		"\n",
		name, "status:\n",
	)
	printMetadata(metadata)

	return
}

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

// pollHome is a loop asking users for questions about managing participants.
func pollServer(c *client.Client) {
	var input string
	var err error
	for {
		fmt.Println()
		fmt.Print("(Server Mode) Please enter a command: ")
		_, err = fmt.Scanln(&input)
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
			err = joinQuorumWalkthrough(c)

		case "n", "new", "bootstrap":
			err = newQuorumWalkthrough(c)

		case "p", "print", "status":
			err = serverMetadataWalkthrough(c)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
