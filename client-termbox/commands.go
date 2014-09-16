package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/network"

	"github.com/spf13/cobra"
)

func commandStart(cmd *cobra.Command, args []string) {
	var err error
	networkState.Router, err = network.NewRPCServer(config.Port)
	if err != nil {
		fmt.Println(err)
		return
	}

	networkState.ServerAddress = network.Address{
		Host: config.ServerHostname,
		Port: config.ServerPort,
		ID:   network.Identifier(config.ServerID),
	}

	termboxRun()
}

func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Termbox Client v0.0.2.1")
}

func manageCommands() {
	root := &cobra.Command{
		Use:   "client-termbox",
		Short: "Sia Client v0.0.2.1",
		Long:  "Sia Client, for interacting with the Sia Server",
		Run:   commandStart,
	}

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Prints version information about the Sia Termbox Client",
		Run:   version,
	}

	root.Flags().Uint16VarP(&config.Port, "port", "p", 9980, "Which port the client messenger should listen on")
	root.Flags().StringVarP(&config.ServerHostname, "server-hostname", "H", "localhost", "The hostname of the server you are connecting to.")
	root.Flags().Uint16VarP(&config.ServerPort, "server-port", "P", 9988, "The port on which the server is listening.")
	root.Flags().Uint8VarP(&config.ServerID, "server-id", "I", 1, "The id of the server you are connecting to.")

	root.AddCommand(version)
	root.Execute()
}
