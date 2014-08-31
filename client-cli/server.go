package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/client"
	"github.com/NebulousLabs/Sia/state"
)

// participantName is a helper function that asks for a participant name.
func participantName() (name string, err error) {
	fmt.Print("Name of participant: ")
	_, err = fmt.Scanln(&name)
	if err != nil {
		return
	}

	return
}

// printMetadata is a helper function that takes metadata and outputs it in
// human readable form.
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

// printMetadataVerbose is a helper function that takes metadata and outputs it in
// human readable form, providing much more information.
func printMetadataVerbose(metadata state.Metadata) {
	var siblingString string
	for i := range metadata.Siblings {
		if metadata.Siblings[i].Active() {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Active\n", siblingString, i)
		} else if metadata.Siblings[i].Inactive() {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Inactive\n", siblingString, i)
		} else {
			siblingString = fmt.Sprintf("%s\t\tSibling %v: Passive for %v more compiles.", siblingString, i, metadata.Siblings[i].Status)
		}
		siblingString = fmt.Sprintf("%s\t\t\tFull Data: %v", siblingString, metadata.Siblings[i])
	}

	fmt.Println(
		"\tSiblings: \n",
		siblingString,
		"\tEvent Counter:", metadata.EventCounter, "\n",
		"\tStorage Price:", metadata.StoragePrice, "\n",
		"\tParent Block:", metadata.ParentBlock, "\n",
		"\tHeight:", metadata.Height, "\n",
		"\tRecent Snapshot:", metadata.RecentSnapshot, "\n",
		"\tGerm:", metadata.Germ, "\n",
		"\tSeed:", metadata.Seed, "\n",
		"\tPoStorageSeed", metadata.PoStorageSeed, "\n",
	)
}

// printWalletList takes a slice of wallets and prints each in human readable
// form.
func printWalletList(wallets []state.Wallet) {
	var walletString string
	for _, wallet := range wallets {
		walletString += fmt.Sprintf("\tWallet: %v\n", wallet.ID)
		walletString += fmt.Sprintf("\t\tBalance: %v\n", wallet.Balance)
		walletString += fmt.Sprintf("\t\tSector:\n")
		{
			walletString += fmt.Sprintf("\t\t\tAtoms: %v\n", wallet.Sector.Atoms)
			walletString += fmt.Sprintf("\t\t\tK: %v\n", wallet.Sector.K)
			walletString += fmt.Sprintf("\t\t\tD: %v\n", wallet.Sector.D)
			walletString += fmt.Sprintf("\t\t\tHash: %v\n", wallet.Sector.Hash())
		}
		walletString += fmt.Sprintf("\t\tScript: %v", wallet.Script)
		walletString += "\n\n"
	}

	fmt.Println(walletString)
}

// printWalletListVerbose takes a slice of wallets and prints each in human readable
// form, providing much more information.
func printWalletListVerbose(wallets []state.Wallet) {
	var walletString string
	for _, wallet := range wallets {
		walletString += fmt.Sprintf("\tWallet: %v\n", wallet.ID)
		walletString += fmt.Sprintf("\t\tBalance: %v\n", wallet.Balance)
		walletString += fmt.Sprintf("\t\tSector:\n")
		{
			walletString += fmt.Sprintf("\t\t\tAtoms: %v\n", wallet.Sector.Atoms)
			walletString += fmt.Sprintf("\t\t\tK: %v\n", wallet.Sector.K)
			walletString += fmt.Sprintf("\t\t\tD: %v\n", wallet.Sector.D)
			walletString += fmt.Sprintf("\t\t\tHashSet: %v\n", wallet.Sector.HashSet)
			walletString += fmt.Sprintf("\t\t\tActiveUpdates: %v\n", wallet.Sector.ActiveUpdates)
		}
		walletString += fmt.Sprintf("\t\tScript: %v", wallet.Script)
		walletString += fmt.Sprintf("\t\tKnown Scripts: %v", wallet.KnownScripts)
		walletString += "\n\n"
	}

	fmt.Println(walletString)
}

// serverNameAndFilepathWalkthrough is a helper function that gets and returns
// the name and folderpath of a server.
func serverNameAndFilepathWalkthrough(c *client.Client) (name string, filepath string, err error) {
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
	fmt.Println("Entering 'New Quorum' walkthorugh")
	fmt.Println("Warning: The client you are using was only intended to work with a single network. This function creates a new Sia network. If you have existing wallets, it's possible that there will be problems.")
	fmt.Println("Double Warning: This feature is really only be meant to be used by developers. A lot can go wrong, just please be careful and realized that you were warned if bad stuff happens.")

	name, filepath, err := serverNameAndFilepathWalkthrough(c)
	if err != nil {
		return
	}

	// Establish a wallet for the first participant in the quorum.
	var sibID state.WalletID
	fmt.Print("Now creating a wallet for the participant to use. Pick an id: ")
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

// participantMetadataWalkthrough gets the name of a participant and then
// prints the metadata of that participant.
func participantMetadataWalkthrough(c *client.Client) (err error) {
	name, err := participantName()
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

// Prints the metadata along with all of the wallets.
func participantVerboseWalkthrough(c *client.Client) (err error) {
	name, err := participantName()
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
	printMetadataVerbose(metadata)

	wallets, err := c.ParticipantWallets(name)
	if err != nil {
		return
	}

	fmt.Println(
		"\n",
		name, "wallets:\n",
	)
	printWalletListVerbose(wallets)

	return
}

// participantWalletsWalkthrough gets the name of a participant and prints all
// of the wallets being tracked by that participant.
func participantWalletsWalkthrough(c *client.Client) (err error) {
	name, err := participantName()
	if err != nil {
		return
	}

	wallets, err := c.ParticipantWallets(name)
	if err != nil {
		return
	}

	fmt.Println(
		"\n",
		name, "wallets:\n",
	)
	printWalletList(wallets)

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

func displayServerHelp() {
	fmt.Println(
		" h:\tHelp\n",
		"q:\tReturn to home mode.\n",
		"j:\tCreate a new participant and join an existing quorum.\n",
		"m:\tPrint the metadata of a participant.\n",
		"n:\tCreate a new quorum with a bootstrap participant.\n",
		"w:\tPrint the wallets known to a participant.\n",
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
		fmt.Println()

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "help":
			displayServerHelp()

		case "q", "quit", "return":
			return

		case "j", "join":
			err = joinQuorumWalkthrough(c)

		case "m", "metadata", "status":
			err = participantMetadataWalkthrough(c)

		case "n", "new", "bootstrap":
			err = newQuorumWalkthrough(c)

		case "v", "verbose":
			err = participantVerboseWalkthrough(c)

		case "w", "wallets":
			err = participantWalletsWalkthrough(c)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			err = nil
		}
	}
}
