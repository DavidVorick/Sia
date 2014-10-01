package main

import (
	"fmt"
)

// SettingsView is an InputsView that allows the user to modify client
// settings.
type SettingsView struct {
	InputsView
	clientPort string
	serverHost string
	serverPort string
	serverID   string
}

func newSettingsView(parent View) *SettingsView {
	sv := new(SettingsView)

	// load current config values
	sv.clientPort = fmt.Sprint(config.Client.Port)
	sv.serverHost = config.Server.Host
	sv.serverPort = fmt.Sprint(config.Server.Port)
	sv.serverID = fmt.Sprint(config.Server.ID)

	sv.inputs = []Input{
		newForm(sv, "Client Port:", &sv.clientPort, 20),
		newForm(sv, "Server Host:", &sv.serverHost, 20),
		newForm(sv, "Server Port:", &sv.serverPort, 20),
		newForm(sv, "Server ID:  ", &sv.serverID, 20),
	}
	sv.offsets = []int{1, 3, 4, 5}
	sv.Parent = parent
	return sv
}
