package main

import (
	"client"
	"fmt"
	"io/ioutil"
	"state"
)

//Two states:
//Start, where you choose a wallet or do stuff not pertaining to a specific wallet
//Wallet, where you do operations specific to that wallet

func displayHelpStart() {
	fmt.Println("\nc:\tConnect to Network\n" +
		"w:\tRequest new wallet\n" +
		"l:\tList stored wallets\n" +
		"s:\tSave all wallets\n" +
		"e:\tEnter wallet\n")
}

func displayHelpWallet() {
	fmt.Println("\nd:\tDownload a file\n" +
		"r:\tResize a sector\n" +
		"s:\tSend a custom script input\n" +
		"t:\tSubmit transaction\n" +
		"u:\tUpload a file\n")
}

func connect(c *client.Client) {
	// if we are already connected, disconnect first
	c.Disconnect()
	var host string
	var port, id int
	fmt.Print("Hostname: ")
	_, err := fmt.Scanf("%s", &host)
	if err != nil {
		fmt.Println("Invalid hostname")
		return
	}
	fmt.Print("Port: ")
	_, err = fmt.Scanf("%d", &port)
	if err != nil {
		fmt.Println("Invalid port")
		return
	}
	fmt.Print("ID: ")
	_, err = fmt.Scanf("%d", &id)
	if err != nil {
		fmt.Println("Invalid id")
		return
	}
	err = c.Connect(host, port, id)
	if err != nil {
		fmt.Println("Error while connecting:", err)
	} else {
		fmt.Println("Successfully Connected!")
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

func enterWallet(c *client.Client) {
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
	for {
		fmt.Printf("Loaded into wallet #%x\n", c.CurID)
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

func pollStartActions(c *client.Client) {
	var input string
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "?", "h", "help":
			displayHelpStart()

		case "c", "conncet":
			connect(c)

		case "w", "wallet":
			createGenericWallet(c)

		case "l", "ls", "list":
			listWallets(c)

		case "s", "save":
			c.SaveAllWallets()

		case "e", "enter":
			enterWallet(c)

		case "q", "quit":
			return
		}
		input = ""
	}
}

func main() {
	fmt.Println("Sia Client Version 0.0.1")
	c, err := client.NewClient()
	if err == nil {
		fmt.Println("Connected to local bootstrap")
	} else {
		fmt.Println("Autoconnect failed: press c to connect manually")
	}
	if c.CurID != 0 {
		pollWalletActions(c)
	} else {
		pollStartActions(c)
	}
}
