package main

import (
	"fmt"
)

type SettingsView struct {
	InputView
}

func newSettingsView(parent View) View {
	// convert config values to strings
	clientPort := fmt.Sprint(config.Client.Port)
	serverPort := fmt.Sprint(config.Server.Port)
	serverID := fmt.Sprint(config.Server.ID)

	sv := new(SettingsView)
	sv.settings = []*Setting{
		{Field: Field{text: clientPort}, name: "Client Port:", width: 20, offset: 1},
		{Field: Field{text: config.Server.Host}, name: "Server Host:", width: 20, offset: 3},
		{Field: Field{text: serverPort}, name: "Server Port:", width: 20, offset: 4},
		{Field: Field{text: serverID}, name: "Server ID:  ", width: 20, offset: 5},
	}
	sv.Parent = parent
	for _, s := range sv.settings {
		s.Parent = sv
		s.color = SettingColor
	}
	return sv
}
