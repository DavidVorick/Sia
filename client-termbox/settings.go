package main

import (
	"fmt"
	"strconv"
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
		newButton(s, "Save", s.save),
	}
	s.offsets = []int{1, 3, 4, 5, 7}
	s.Parent = parent
	return s
}

// save settings to global Config and config file
func (s *SettingsMVC) save() {
	// validate values
	cport, err := strconv.Atoi(s.clientPort)
	if err != nil || cport < 1024 || cport > 65535 {
		//drawError("invalid client port number")
		return
	}
	// TODO: warn if host not pingable?
	if s.serverHost == "" {
		//drawError("invalid hostname")
		return
	}
	sport, err := strconv.Atoi(s.serverPort)
	if err != nil || sport < 1024 || sport > 65535 {
		//drawError("invalid server port number")
		return
	}
	sid, err := strconv.Atoi(s.serverID)
	if err != nil || sid > 255 {
		//drawError("invalid server ID")
		return
	}

	config.Client.Port = uint16(cport)
	config.Server.Host = s.serverHost
	config.Server.Port = uint16(sport)
	config.Server.ID = byte(sid)

	server.UpdateAddress()

	// write values to config file
	// NOTE: this could also be implemented as a method of Config. However, for
	// practical reasons it is simpler to implement it here, as the values are
	// already in string form. This will be changed if write support is added
	// to the gcfg package.

	//drawInfo("saved to " + configFilename + "!")
}
