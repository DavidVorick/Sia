package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/network"

	"github.com/spf13/cobra"
)

type Config struct {
	Port   uint16
	Server network.Address
	Router *network.RPCServer
}

// global config variable
var config Config

func commandStart(cmd *cobra.Command, args []string) {
	var err error
	config.Router, err = network.NewRPCServer(config.Port)
	if err != nil {
		fmt.Println(err)
		return
	}

	termboxRun()
}

func main() {
	root := &cobra.Command{
		Use:   "client-termbox",
		Short: "Sia Client v0.0.2.1",
		Long:  "Sia Client, for interacting with the Sia Server",
		Run:   commandStart,
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Prints version information about the Sia Termbox Client",
		Run:   func(_ *cobra.Command, _ []string) { fmt.Println("Sia Termbox Client v0.0.2.1") },
	})

	root.Flags().Uint16VarP(&config.Port, "port", "p", 9980, "Which port the client messenger should listen on")
	root.Flags().StringVarP(&config.Server.Host, "server-hostname", "H", "localhost", "The hostname of the server you are connecting to.")
	root.Flags().Uint16VarP(&config.Server.Port, "server-port", "P", 9988, "The port on which the server is listening.")
	var id uint8
	root.Flags().Uint8VarP(&id, "server-id", "I", 1, "The id of the server you are connecting to.")
	config.Server.ID = network.Identifier(id)

	root.Execute()
}
