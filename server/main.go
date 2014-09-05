package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/spf13/cobra"
)

var configLocation string
var port uint16
var publicConnection bool

// defaultConfigLocation checks a bunch of places for the config file, in a
// particular order, and then returns the first one found.
func defaultConfigLocation() (configLocation string) {
	// Check the home directory for a config file.
	homeLocation, err := siafiles.HomeFilename("config-server")
	if err != nil {
		// Log something, but somewhere that normal users will never see.
	} else {
		if siafiles.Exists(homeLocation) {
			configLocation = homeLocation
			return
		}
	}

	// Check the /etc directory for a config file... not sure what the
	// windows equivalent would be.
	rootLocation := "/etc/siaserver-config"
	if siafiles.Exists(rootLocation) {
		configLocation = rootLocation
		return
	}

	return
}

// start takes all the values parsed by the flags and config files, and creates
// the server.
func start(cmd *cobra.Command, args []string) {
	// Initialize the server.
	fmt.Println("Starting Sia Server...")
	s := newServer()

	// Connect the server, which will prepare it to listen for rpc's.
	s.connect(port, publicConnection)

	// Let the server run indefinitely.
	for {
	}
}

// version prints version information about the server.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Server v0.0.3")
}

// main uses cobra to parse commands and flags.
func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "Sia server v0.0.3",
		Long:  "Sia server, meant to run in the background and interface with a client. Version 0.0.3",
		Run:   start,
	}

	// Search for a config file, and use that as the default.
	dcl := defaultConfigLocation()
	root.Flags().StringVarP(&configLocation, "config", "c", dcl, "Where to find the server configuration file.")
	// parse the config file into a struct

	// Load the config file into a struct, use the struct to see if a default port was set.
	// Set the default port to the value specified by the default config file.
	defaultPort := uint16(9988)
	root.Flags().Uint16VarP(&port, "port", "p", defaultPort, "Which port the server should listen on.")

	// Flag for determining if the server should be local or public.
	root.Flags().BoolVarP(&publicConnection, "public", "P", false, "Set this flag to have a publically visible hostname.")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   version,
	}

	root.AddCommand(version)
	root.Execute()
}
