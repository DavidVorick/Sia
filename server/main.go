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
		Short: "Sia server",
		Long:  "Sia server, meant to run in the background and interface with a client.",
		Run:   start,
	}

	dcl := defaultConfigLocation()
	var configLocation string
	root.Flags().StringVarP(&configLocation, "config", "c", dcl, "Where to find the server configuration file.")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   version,
	}

	root.AddCommand(version)

	root.Execute()
}
