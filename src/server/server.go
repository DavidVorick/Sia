package server

import (
	"network"
)

// The Server houses all of the participants. It contains a single message
// router that is shared by all of the participants, it will eventually contain
// a clock object that will be used and modified by all participants.
type Server struct {
	rpcServer *network.RPCServer
}

// NewServer takes a port number as input and returns a server object that's
// ready to be populated with participants.
func NewServer(port int) (s *Server, err error) {
	s = new(Server)

	// Create the RPCServer
	s.rpcServer, err = network.NewRPCServer(port)
	if err != nil {
		return
	}
	return
}

/*
func joinQuorum() {
	// read and set bootstrap address
	var bootstrap network.Address
	fmt.Print("Bootstrap hostname: ")
	fmt.Scanf("%s", &bootstrap.Host)
	fmt.Print("Bootstrap port: ")
	fmt.Scanf("%d", &bootstrap.Port)
	fmt.Print("Bootstrap ID: ")
	fmt.Scanf("%d", &bootstrap.ID)
	err = networkServer.Ping(bootstrap)
	if err != nil {
		fmt.Println("Failed to ping bootstrap:", err)
		return
	}

	// read wallet ID
	var sibID uint64
	fmt.Print("Wallet ID: ")
	fmt.Scanf("%d", &sibID)

	// obtain a wallet
	err = networkServer.SendMessage(network.Message{
		Dest: bootstrap,
		Proc: "Participant.AddScriptInput",
		Args: delta.CreateWalletInput(sibID, []byte{0x2F}), // need better default script
		Resp: nil,
	})
	if err != nil {
		fmt.Println("Failed to obtain wallet:", err)
		return
	}

	var directory string
	fmt.Print("participant directory: ")
	fmt.Scanf("%s", &directory)

	// create a participant
	_, err = consensus.CreateJoiningParticipant(networkServer, directory, state.WalletID(sibID), []network.Address{bootstrap})
	if err != nil {
		fmt.Println(err)
		return
	}

	// block forever
	select {}
}

func establishQuorum() {
	// read and set port number
	var port int
	fmt.Print("Port to listen on: ")
	fmt.Scanf("%d", &port)

	// create a message router
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer networkServer.Close()

	var directory string
	fmt.Print("participant directory: ")
	fmt.Scanf("%s", &directory)

	var sibID state.WalletID
	fmt.Print("Wallet ID: ")
	fmt.Scanf("%d", &sibID)

	// create a participant
	_, err = consensus.CreateBootstrapParticipant(networkServer, directory, sibID)
	if err != nil {
		fmt.Println(err)
		return
	}
	select {}
}
*/
