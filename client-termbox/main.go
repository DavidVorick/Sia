package main

type Config struct {
	Port uint16

	ServerHostname string
	ServerPort     uint16
	ServerID       uint8
}

var config Config

func main() {
	manageCommands()
}
