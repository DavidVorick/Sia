package main

import (
	"client"
	"fmt"
	"network"
	"participant"
	"quorum"
	"quorum/script"
	"siacrypto"
)

var (
	router *network.RPCServer
)

// request a new wallet from the bootstrap
func requestWallet(id quorum.WalletID) error {
	return router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: participant.BootstrapID,
			Input:    script.CreateWalletInput(uint64(id), script.TransactionScript),
		},
		Resp: nil,
	})
}

// send coins from one wallet to another
func submitTransaction(src, dst quorum.WalletID, amount uint64) (err error) {
	return router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: src,
			Input:    script.TransactionInput(uint64(dst), 0, amount),
		},
		Resp: nil,
	})
}

// resize sector associated with wallet
func resizeSector(w quorum.WalletID, atoms uint16, m byte) (err error) {
	return router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: w,
			Input:    script.ResizeSectorEraseInput(atoms, m),
		},
		Resp: nil,
	})
}

func connectToBootstrap() (err error) {
	router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}
	err = router.Ping(&participant.BootstrapAddress)
	return
}

func main() {
	// input values
	var (
		input, destFile   string
		err               error
		id, srcID, destID quorum.WalletID
		amount            uint64
		atoms             uint16
		m                 byte
	)
	fmt.Println("Sia Client Version 0.0.0.3")
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "help":
			fmt.Println("c:\tConnect to bootstrap\n" +
				"w:\tRequest wallet\n" +
				"t:\tSubmit transaction\n" +
				"g:\tGenerate public and secret key pair\n" +
				"r:\tResize a sector")

		case "c":
			err = connectToBootstrap()
			if err != nil {
				fmt.Println("Could not connect to bootstrap:", err)
			} else {
				fmt.Println("Connected to bootstrap")
			}

		case "w":
			fmt.Print("Enter desired Wallet ID (hex): ")
			fmt.Scanf("%x", &id)
			err = requestWallet(id)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Wallet requested")
			}

		case "t":
			fmt.Print("Source Wallet ID (hex): ")
			fmt.Scanf("%x", &srcID)
			fmt.Print("Dest Wallet ID (hex): ")
			fmt.Scanf("%x", &destID)
			fmt.Print("Amount to send (dec): ")
			fmt.Scanln(&amount)
			err = submitTransaction(srcID, destID, amount)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Transaction successfully submitted")
			}

		case "g":
			publicKey, secretKey, err := siacrypto.CreateKeyPair()
			if err != nil {
				panic(err)
			}
			fmt.Println("keys generated. Where would you like to store them? ")
			fmt.Scanf("%s", &destFile)
			fmt.Println("Saving to:", destFile)
			err = client.SaveKeyPair(publicKey, secretKey, destFile)
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Success!")
			}

		case "r":
			fmt.Print("Wallet ID (hex): ")
			fmt.Scanf("%x", &srcID)
			fmt.Print("New size (in atoms): ")
			fmt.Scanln(&atoms)
			fmt.Print("Redundancy: ")
			fmt.Scanln(&m)
			err = resizeSector(srcID, atoms, m)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Sector resized")
			}

		case "u":
			var filename string
			fmt.Print("Filename: ")
			fmt.Scanln(&filename)
			fmt.Print("M: ")
			fmt.Scanln(&m)
			atomsRequired, err := client.CalculateAtoms(filename, int(m))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Atoms Required: %v\n", atomsRequired)
			}

		case "q":
			return
		}
	}
}
