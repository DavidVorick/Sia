package main

import (
	"fmt"
)

// SettingsMVC allows the user to modify client settings.
type SettingsMVC struct {
	InputGroupMVC
	clientPort string
	serverHost string
	serverPort string
	serverID   string
}

func newSettingsMVC(parent MVC) *SettingsMVC {
	s := new(SettingsMVC)

	// load current config values
	s.clientPort = fmt.Sprint(config.Client.Port)
	s.serverHost = config.Server.Host
	s.serverPort = fmt.Sprint(config.Server.Port)
	s.serverID = fmt.Sprint(config.Server.ID)

	s.inputs = []Input{
		newForm(s, "Client Port:", &s.clientPort, 20),
		newForm(s, "Server Host:", &s.serverHost, 20),
		newForm(s, "Server Port:", &s.serverPort, 20),
		newForm(s, "Server ID:  ", &s.serverID, 20),
	}
	s.offsets = []int{1, 3, 4, 5}
	s.Parent = parent
	return s
}
