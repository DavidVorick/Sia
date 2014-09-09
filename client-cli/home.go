package main

/*
import (
	"errors"
	"fmt"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/server"
	"github.com/NebulousLabs/Sia/state"
)

// printWallets provides a list of every wallet available to the Client.
func printWallets(s *server.Server) {
	fmt.Println("All Stored Wallet IDs:")
	wallets := s.GetWalletIDs()
	for _, id := range wallets {
		fmt.Printf("%x\n", id)
	}
}

// connectWalkthrough guides the user through providing a hostname, port, and
// id which can be used to create a Sia address. Then the connection is
// committed.
func bootstrapToNetworkWalkthrough(s *server.Server) (err error) {
	fmt.Println("Starting connect walkthrough.")
	if !s.IsRouterInitialized() {
		err = connectWalkthrough(s)
		if err != nil {
			return
		}
	}

	fmt.Println("Please indicate the hostname, port, and id that you wish to connect through.")
	var connectAddress network.Address

	// Load the hostname.
	fmt.Print("Hostname: ")
	_, err = fmt.Scanln(&connectAddress.Host)
	if err != nil {
		return
	}

	// Load the port number.
	fmt.Print("Port: ")
	_, err = fmt.Scanln(&connectAddress.Port)
	if err != nil {
		return
	}

	// Load the participant id.
	fmt.Print("ID: ")
	_, err = fmt.Scanln(&connectAddress.ID)
	if err != nil {
		return
	}

	// Call server.Connect using the provided information.
	err = s.BootstrapConnection(connectAddress)
	if err != nil {
		return
	} else {
		fmt.Println("Connection successful.")
	}

	return
}

// connectWalkthrough requests a port and then calls server.Connect(port),
// initializing the server network router.
func connectWalkthrough(s *server.Server) (err error) {
	// Do nothing if the router is already initialized.
	if s.IsRouterInitialized() {
		err = errors.New("router is already initialized")
		return
	}

	// Get a port.
	var port uint16
	fmt.Print("Port the server should listen on: ")
	_, err = fmt.Scanln(&port)
	if err != nil {
		err = errors.New("invalid port")
		return
	}

	err = s.Connect(port)
	return
}

func createGenericWalletWalkthrough(s *server.Server) (err error) {
	var id state.WalletID
	fmt.Print("Enter desired Wallet ID: ")
	_, err = fmt.Scanln(&id)
	if err != nil {
		return
	}

	errChan := make(chan error)
	go func() {
		err := s.RequestGenericWallet(id)
		errChan <- err
	}()

	for i := byte(0); i < consensus.NumSteps*2; i++ {
		time.Sleep(consensus.StepDuration)
		fmt.Printf("Step %v of %v\n", i, consensus.NumSteps*2)
	}

	err = <-errChan
	if err != nil {
		return
	}

	fmt.Println("Wallet Recieved")

	return
}

// loadWallet switches the cli into wallet-mode, where actions are taken
// against a specific wallet.
func loadWalletWalkthrough(s *server.Server) (err error) {
	// Fetch the wallet id from the user.
	var id state.WalletID
	fmt.Print("Wallet ID: ")
	_, err = fmt.Scanln(&id)
	if err != nil {
		return
	}

	// Check that the wallet is available to the server.
	walletType, err := s.WalletType(id)
	if err != nil {
		return
	}

	// If the wallet type is recognized, switch to the polling of that wallet
	// type. If the type is not recognized, print an error and return to the
	// home menu.
	if err != nil {
		return
	} else if walletType == "generic" {
		var gw server.GenericWallet
		gw, err = s.GenericWallet(server.GenericWalletID(id))
		if err != nil {
			return
		}
		pollGenericWallet(s, gw)
	} else {
		err = errors.New("wallet is available, but is of an unknown type.")
		return
	}

	return
}

// serverModeSwitch will transition the server from being in home mode to being
// in server mode, creating a new server and a new router if necessary.
func serverModeSwitch(s *server.Server) (err error) {
	init := s.IsParticipantManagerInitialized()
	if !init {
		err = serverCreationWalkthrough(s)
		if err != nil {
			return
		}
	}
	pollServer(s)
	return
}

// displayHomeHelp lists all of the options available to the home screen, with
// a description of what they do.
func displayHomeHelp() {
	fmt.Println(
		" h:\tHelp\n",
		"q:\tQuit\n",
		"c:\tConnect to the network using a bootstrap address\n",
		"g:\tRequest a new generic wallet\n",
		"l:\tLoad wallet\n",
		"p:\tPrint wallets\n",
		"s:\tSwitch to server mode, creating a server if none yet exists.\n",
		"S:\tSave all wallets\n",
	)
}

// pollHome maintains the loop that asks users for actions that are relevant to the home screen.
func pollHome(s *server.Server) {
	var input string
	var err error
	for {
		fmt.Println()
		fmt.Print("(Home) Please enter a command: ")
		_, err = fmt.Scanln(&input)
		if err != nil {
			continue
		}
		fmt.Println()

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHomeHelp()

		case "q", "quit":
			return

		case "b", "bootstrap", "c", "connect":
			err = bootstrapToNetworkWalkthrough(s)

		case "g", "generic", "request", "new":
			err = createGenericWalletWalkthrough(s)

		case "l", "load", "enter":
			err = loadWalletWalkthrough(s)

		case "p", "ls", "print", "list":
			printWallets(s)

		case "s", "server":
			err = serverModeSwitch(s)

		case "S", "save":
			fmt.Println("Saving all wallets...")
			s.SaveAllWallets()
			fmt.Println("...finished!")
		}

		if err != nil {
			fmt.Println("Error:", err)
			err = nil
		}
	}
}
*/
