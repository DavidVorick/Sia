package main

import (
	"fmt"
)

type SettingsView struct {
	InputsView
}

func newSettingsView(parent View) View {
	// convert config values to strings
	clientPort := fmt.Sprint(config.Client.Port)
	serverHost := config.Server.Host
	serverPort := fmt.Sprint(config.Server.Port)
	serverID := fmt.Sprint(config.Server.ID)

	sv := new(SettingsView)
	sv.inputs = []Input{
		newForm(sv, "Client Port:", clientPort, 20, 1),
		newForm(sv, "Server Host:", serverHost, 20, 3),
		newForm(sv, "Server Port:", serverPort, 20, 4),
		newForm(sv, "Server ID:  ", serverID, 20, 5),
	}
	sv.Parent = parent
	return sv
}
