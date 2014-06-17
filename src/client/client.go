package main

import (
	"fmt"
)

// ah fudge we're going to need a way to load public keys
func submitTransaction() {
	var srcID uint64
	var destID uint64
	var amount uint64
	print("Source Wallet ID (hex): ")
	fmt.Scanf("%x", &srcID)
	print("Dest Wallet ID (hex): ")
	fmt.Scanf("%x", &destID)
	print("Amount to send (dec): ")
	fmt.Scanf("%d", &amount)
	println()
	fmt.Printf("Sending %v siacoins from wallet %x to wallet %x, y/n: ", amount, srcID, destID)

	println("Luke will implement something here :)")
}

func main() {
	var input string
	for {
		// get some input
		println("Sia Client Version 0.0.0.1")
		print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			println("unrecognized command")

		case "t":
			submitTransaction()

		case "q":
			break
		}
	}
}
