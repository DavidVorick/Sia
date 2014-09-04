package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/siafiles"

	"github.com/spf13/cobra"
)

// defaultConfigLocation checks a bunch of places for the config file, in a
// particular order, and then returns the first one found.
func defaultConfigLocation() (configLocation string) {
	// First check the home directory.
	// homeLocation := path.Join(siafiles.Home(), ".config", "Sia", "config-server")
	// if siafiles.Exists(homelocation) {
	// 	return
	// }
	//
	// homeLocation2 := path.Join(siafiles.Home(), ".Sia", "config-server")
	// etc.

	// Really need some way to get the root location through siafiles too... this isn't Windows friendly
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
