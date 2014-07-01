package main

import (
	"client"
	"fmt"
	"quorum"
)

func displayHelp() {
	fmt.Println("\nc:\tConnect to Network\n" +
		"d:\tDownload a file\n" +
		"r:\tResize a sector\n" +
		"t:\tSubmit transaction\n" +
		"u:\tUpload a file\n" +
		"w:\tRequest wallet\n")
}

func connect(c *client.Client) {
	var host string
	var port int
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
	err = c.Connect(host, port)
	if err != nil {
		fmt.Println("Error while connecting:", err)
	} else {
		fmt.Println("Successfully Connected!")
	}
}

func download(c *client.Client) {
	var dest string
	var id quorum.WalletID
	fmt.Print("Wallet ID (hex): ")
	fmt.Scanf("%x", &id)
	fmt.Print("Destination Filepath: ")
	fmt.Scanln(&dest)
	fmt.Println("Downloading File, please wait a few minutes")
	c.Download(id, dest)
}

func resizeGenericWallet(c *client.Client) {
	var srcID quorum.WalletID
	var atoms uint16
	var m byte
	fmt.Print("Wallet ID (hex): ")
	fmt.Scanf("%x", &srcID)
	fmt.Print("New size (in atoms): ")
	fmt.Scanln(&atoms)
	fmt.Print("Redundancy: ")
	fmt.Scanln(&m)
	err := c.ResizeSector(srcID, atoms, m)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Sector resized")
	}
}

func sendFromGenericWallet(c *client.Client) {
	var srcID quorum.WalletID
	var destID quorum.WalletID
	var amount uint64
	fmt.Print("Source Wallet ID (hex): ")
	fmt.Scanf("%x", &srcID)
	fmt.Print("Dest Wallet ID (hex): ")
	fmt.Scanf("%x", &destID)
	fmt.Print("Amount to send (dec): ")
	fmt.Scanln(&amount)
	err := c.SubmitTransaction(srcID, destID, amount)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Transaction successfully submitted")
	}
}

func uploadToGenericWallet(c *client.Client) {
	var filename string
	var id quorum.WalletID
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
	fmt.Print("Wallet ID (hex): ")
	fmt.Scanf("%x", &id)
	fmt.Println("Attempting to Upload File, please wait a few minutes (longer for large files).")
	c.UploadFile(id, filename, k)
}

func createGenericWallet(c *client.Client) {
	var id quorum.WalletID
	fmt.Print("Enter desired Wallet ID (hex): ")
	fmt.Scanf("%x", &id)
	err := c.RequestWallet(id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Wallet requested")
	}
}

func main() {
	fmt.Println("Sia Client Version 0.0.0.3")
	c, err := client.NewClient()
	if err == nil {
		fmt.Println("Connected to local bootstrap")
	} else {
		fmt.Println("Autoconnect failed: press c to connect manually")
	}

	var input string
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "help":
			displayHelp()

		case "c", "conncet":
			connect(c)

		case "d", "download":
			download(c)

		case "r", "resize":
			resizeGenericWallet(c)

		case "t", "send", "transaction":
			sendFromGenericWallet(c)

		case "u", "upload":
			uploadToGenericWallet(c)

		case "w", "wallet":
			createGenericWallet(c)

		case "q", "quit":
			return
		}
	}
}
