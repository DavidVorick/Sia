package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siafiles"

	"code.google.com/p/gcfg"
	"github.com/spf13/cobra"
)

// Config resembles the structure of the client config file.
type Config struct {
	Client struct {
		Port uint16
	}
	Server struct {
		Host string
		Port uint16
		ID   byte
	}
}

// global config variable
var config Config
var configFilename string

// global server variable
var server Server

func init() {
	homeLocation, err := siafiles.HomeFilename("config-client")
	if err == nil && siafiles.Exists(homeLocation) {
		configFilename = homeLocation
	} else if etcLocation := "/etc/config-client"; siafiles.Exists(etcLocation) {
		configFilename = etcLocation
	}
}

func commandStart(cmd *cobra.Command, args []string) {
	var err error
	server.Router, err = network.NewRPCServer(config.Client.Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	server.UpdateAddress()
	// TODO: check that server is reachable
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

	// The config data is initially set here. These defaults have the lowest priority.
	root.Flags().Uint16VarP(&config.Client.Port, "port", "p", 9980, "Which port the client messenger should listen on")
	root.Flags().StringVarP(&config.Server.Host, "server-hostname", "H", "localhost", "The hostname of the server you are connecting to.")
	root.Flags().Uint16VarP(&config.Server.Port, "server-port", "P", 9988, "The port on which the server is listening.")
	root.Flags().Uint8VarP(&config.Server.ID, "server-id", "I", 1, "The id of the server you are connecting to.")

	// Then, config file is parsed, and any modified values are overwritten.
	if configFilename != "" {
		gcfg.ReadFileInto(&config, configFilename)
		// log error?
	}

	// Finally, the command line arguments are parsed. They have the highest priority.
	root.Execute()
}
