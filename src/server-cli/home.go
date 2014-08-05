package main

import (
	"fmt"
	"server"
)

func displayHomeHelp() {
	fmt.Println(
		"\n",
		"h:\tHelp\n",
		"q:\tQuit\n",
		"j:\tCreate a new participant and join an existing quorum.\n",
		"n:\tCreate a new quorum with a bootstrap participant.\n",
	)
}

// joinQuorumWalkthrough gets input about the bootstrap address, the file
// prefix for the particpant, etc. Then the participant is created and
// bootstrapped to an existing quorum.
func joinQuorumWalkthrough(s *server.Server) {
	// Request and load the bootstrap address.
}

// pollHome is a loop asking users for questions about managing participants.
func pollHome(s *server.Server) {
	var input string
	for {
		input = ""
		fmt.Print("Please enter a command: ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		switch input {
		default:
			fmt.Println("unrecognized command")

		case "h", "help":
			displayHomeHelp()

		case "q":
			return

		case "j":
			joinQuorumWalkthrough(s)

		case "n":
			fmt.Println("This feature has not been implemented.")
			//newQuorumWalkthrough(s)
		}
	}
}
