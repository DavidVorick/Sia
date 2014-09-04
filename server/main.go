package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Server v0.0.3")
}

func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "Sia server",
		Long:  "Sia server, meant to run in the background and interface with a client.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World!")
		},
	}

	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   version,
	}

	root.AddCommand(version)

	root.Execute()
}
