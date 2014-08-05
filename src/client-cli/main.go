package main

import (
	"client"
	"fmt"
	"io/ioutil"
	"state"
)

func printWelcomeMessage() {
	fmt.Println("Sia Client Version 0.0.2")
	fmt.Println("To Connect to the network, press 'c'.")
}

//Two states:
//Start, where you choose a wallet or do stuff not pertaining to a specific wallet
//Wallet, where you do operations specific to that wallet

func displayHelpWallet() {
	fmt.Println("\nd:\tDownload a file\n" +
		"r:\tResize a sector\n" +
		"s:\tSend a custom script input\n" +
		"t:\tSubmit transaction\n" +
		"u:\tUpload a file\n")
}

// connectWalkthrough guides the user through providing a hostname, port, and
// id which can be used to create a Sia address. Then the connection is
// committed.
func connectWalkthrough(c *client.Client) {
	// Eliminate all existing connections.
	c.Disconnect()

	fmt.Println("Please indicate the hostname, port, and id that you wish to connect to.")

	// Load the hostname.
	var host string
	fmt.Print("Hostname: ")
	_, err := fmt.Scanf("%s", &host)
	if err != nil {
		fmt.Println("Invalid hostname")
		return
	}

	// Load the port number.
	var port int
	fmt.Print("Port: ")
	_, err = fmt.Scanf("%d", &port)
	if err != nil {
		fmt.Println("Invalid port")
		return
	}

	// Load the participant id.
	var id int
	fmt.Print("ID: ")
	_, err = fmt.Scanf("%d", &id)
	if err != nil {
		fmt.Println("Invalid id")
		return
	}

	// Call client.Connect using the provided information.
	err = c.Connect(host, port, id)
	if err != nil {
		fmt.Println("Error while connecting:", err)
	} else {
		fmt.Println("Connection successful.")
	}
}

/*
func download(c *client.Client) {
	var dest string
	fmt.Print("Destination Filepath: ")
	fmt.Scanln(&dest)
	fmt.Println("Downloading File, please wait a few minutes")
	c.Download(c.CurID, dest)
}
*/

func resizeGenericWallet(c *client.Client) {
	var atoms uint16
	var m byte
	fmt.Print("New size (in atoms): ")
	fmt.Scanln(&atoms)
	fmt.Print("Redundancy: ")
	fmt.Scanln(&m)
	err := c.ResizeSector(c.CurID, atoms, m)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Sector resized")
	}
}

func sendScriptInput(c *client.Client) {
	var filename string
	fmt.Print("Input file: ")
	fmt.Scanf("%s", &filename)
	// read script from file
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = c.SendCustomInput(c.CurID, input)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Script input sent")
	}
}

func sendFromGenericWallet(c *client.Client) {
	var dstID state.WalletID
	var amount uint64
	fmt.Print("Dest Wallet ID (hex): ")
	fmt.Scanf("%x", &dstID)
	fmt.Print("Amount to send (dec): ")
	fmt.Scanln(&amount)
	err := c.SubmitTransaction(c.CurID, dstID, amount)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Transaction successfully submitted")
	}
}

/*
func uploadToGenericWallet(c *client.Client) {
	var filename string
	var k byte
	fmt.Print("Filename: ")
	fmt.Scanln(&filename)
	fmt.Print("K: ")
	fmt.Scanln(&k)
	atomsRequired, err := client.CalculateAtoms(filename, k)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Atoms Required: %v\n", atomsRequired)
	}
	fmt.Println("Attempting to Upload File, please wait a few minutes (longer for large files).")
	c.UploadFile(c.CurID, filename, k)
}
*/

func createGenericWallet(c *client.Client) {
	var id state.WalletID
	var filename string
	fmt.Print("Enter desired Wallet ID (hex): ")
	fmt.Scanf("%x", &id)
	fmt.Print("Script file (blank for default): ")
	fmt.Scanf("%s", &filename)
	var script []byte
	var err error
	if filename != "" {
		script, err = ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}
	err = c.RequestWallet(id, script)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Wallet requested")
	}
}

func listWallets(c *client.Client) {
	fmt.Println("All Stored Wallet IDs:")
	wallets := c.GetGenericWallets()
	for _, id := range wallets {
		fmt.Printf("%x\n", id)
	}
}

func loadWallet(c *client.Client) {
	var id state.WalletID
	fmt.Print("Wallet ID (hex): ")
	fmt.Scanf("%x", &id)
	err := c.EnterWallet(id)
	if err == nil {
		pollWalletActions(c)
	} else {
		fmt.Println(err)
	}
}

func pollWalletActions(c *client.Client) {
	var input string
	fmt.Printf("Loaded into wallet #%x\n", c.CurID)
	for {
		fmt.Print("Please enter a wallet action: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHelpWallet()

		case "d", "download":
			//download(c)

		case "r", "resize":
			resizeGenericWallet(c)

		case "s", "script":
			sendScriptInput(c)

		case "t", "send", "transaction":
			sendFromGenericWallet(c)

		case "u", "upload":
			//uploadToGenericWallet(c)

		case "q", "quit":
			return
		}
		input = ""
	}
}

func main() {
	printWelcomeMessage()

	c, err := client.NewClient()
	if err != nil {
		fmt.Println("Error on startup:", err)
	}

	pollHome(c)
}
