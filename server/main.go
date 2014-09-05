package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/spf13/cobra"
)

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

func start(cmd *cobra.Command, args []string) {
	fmt.Println("Hello World!")
}

func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Server v0.0.3")
}

func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "Sia server v0.0.3",
		Long:  "Sia server, meant to run in the background and interface with a client. Version 0.0.3",
		Run:   start,
	}

	// Search for a config file, and use that as the default.
	dcl := defaultConfigLocation()
	var configLocation string
	root.Flags().StringVarP(&configLocation, "config", "c", dcl, "Where to find the server configuration file.")

	// Load the config file into a struct, use the struct to see if a default port was set.
	// Set the default port to the value specified by the default config file.
	defaultPort := uint16(9988)

	var port uint16
	root.Flags().Uint16VarP(&port, "port", "p", defaultPort, "Which port the server should listen on.")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   version,
	}

	root.AddCommand(version)
	root.Execute()
}
