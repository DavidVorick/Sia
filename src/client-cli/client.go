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

func connectToBootstrap() (err error) {
	router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}
	err = router.Ping(&participant.BootstrapAddress)
	return
}

func main() {
	var (
		input  string
		err    error
		id     quorum.WalletID
		srcID  quorum.WalletID
		destID quorum.WalletID
		amount uint64
	)
	fmt.Println("Sia Client Version 0.0.0.3")
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")
		case "h", "help":
			fmt.Println()
			fmt.Println("c:\tConnect to bootstrap")
			fmt.Println("w:\tRequest wallet")
			fmt.Println("t:\tSubmit transaction")
			fmt.Println("g:\tGenerate public and secret key pair")
			fmt.Println()
		case "c":
			err = connectToBootstrap()
			if err != nil {
				fmt.Println("Could not connect to bootstrap:", err)
				return
			} else {
				fmt.Println("Connected to bootstrap")
			}

		case "w":
			fmt.Print("Enter desired Wallet ID (hex): ")
			fmt.Scanf("%x", &id)
			err = requestWallet(id)
			if err != nil {
				fmt.Println(err)
				return
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
				return
			} else {
				fmt.Println("Transaction successfully submitted")
			}
		case "g":
			var destFile string
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
		case "q":
			return
		}
	}
}
