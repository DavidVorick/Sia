package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var rootCommand = &cobra.Command{
		Use:   "server",
		Short: "Sia server",
		Long:  "Sia server, meant to run in the background and interface with a client.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World!")
		},
	}

	rootCommand.Execute()
}
