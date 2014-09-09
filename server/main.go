package main

import (
	"fmt"

	"code.google.com/p/gcfg"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/spf13/cobra"
)

type Config struct {
	Network struct {
		Port             uint16
		PublicConnection bool
	}

	Filesystem struct {
		ParticipantDir string
		WalletDir      string
	}
}

var (
	configLocation string
	config         Config
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

// start takes all the values parsed by the flags and config files, and creates
// the server.
func start(cmd *cobra.Command, args []string) {
	// Initialize the server.
	fmt.Println("Starting Sia Server...")
	s := newServer()

	// Connect the server, which will prepare it to listen for rpc's.
	err := s.connect(config.Network.Port, config.Network.PublicConnection)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Operating Variables:")
	fmt.Println("Port:", config.Network.Port)
	fmt.Println("Public Connection:", config.Network.PublicConnection)
	fmt.Println("Participant Directory:", config.Filesystem.ParticipantDir)
	fmt.Println("Wallet Directory:", config.Filesystem.WalletDir)

	// Let the server run indefinitely.
	select {}
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

	// Parse the config file if it exists.
	if siafiles.Exists(configLocation) {
		err := gcfg.ReadFileInto(&config, configLocation)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		config.Network.Port = 9988
		config.Network.PublicConnection = false

		var err error
		config.Filesystem.ParticipantDir, err = siafiles.HomeFilename("participants")
		if err != nil {
			fmt.Println(err)
			return
		}
		config.Filesystem.WalletDir, err = siafiles.HomeFilename("wallets")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Use the config file struct to determine the default port value.  Set
	// the default port to the value specified by the default config file.
	root.Flags().Uint16VarP(&config.Network.Port, "port", "p", config.Network.Port, "Which port the server should listen on.")

	// Flag for determining if the server should be local or public.
	root.Flags().BoolVarP(&config.Network.PublicConnection, "public", "P", config.Network.PublicConnection, "Set this flag to have a publically visible hostname.")

	// Use the config file struct to determine the default participant
	// folder. If none is specified, use the homedir.
	root.Flags().StringVarP(&config.Filesystem.ParticipantDir, "participant-directory", "d", config.Filesystem.ParticipantDir, "Which directory participant files will be saved to.")

	// Use the config file struct to determine the default wallet folder.
	// If none is specified, use the homedir.
	root.Flags().StringVarP(&config.Filesystem.WalletDir, "wallet-directory", "w", config.Filesystem.WalletDir, "Which directory wallets will be loaded from and saved to.")

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   version,
	}

	root.AddCommand(version)
	root.Execute()
}
