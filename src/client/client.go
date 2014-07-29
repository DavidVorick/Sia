package client

import (
	"bufio"
	"errors"
	"fmt"
	"network"
	"os"
	"path/filepath"
	"siacrypto"
	"state"
	"strings"
)

type Keypair struct {
	SK siacrypto.SecretKey
	PK siacrypto.PublicKey
}

// Struct Client contains the state for client actions
type Client struct {
	router         *network.RPCServer
	bootstrap      network.Address
	genericWallets map[state.WalletID]*Keypair
	CurID          state.WalletID
	siblings       [state.QuorumSize]state.Sibling
}

// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (c *Client) Broadcast(nm network.Message) {
	for i := range c.siblings {
		if c.siblings[i].Address.Host == "" {
			continue
		}
		nm.Dest = c.siblings[i].Address
		c.router.SendMessage(nm)
		break
	}
}

// Initializes the client message router and pings the bootstrap to verify
// connectivity.
func (c *Client) Connect(host string, port int, id int) (err error) {
	c.router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}
	// set bootstrap address
	c.bootstrap.Host = host
	c.bootstrap.Port = port
	c.bootstrap.ID = network.Identifier(id)
	err = c.router.Ping(c.bootstrap)
	if err != nil {
		c.router.Close()
		return
	}

	// populate initial sibling list
	var metadata state.StateMetadata
	err = c.router.SendMessage(network.Message{
		Dest: c.bootstrap,
		Proc: "Participant.Metadata",
		Args: struct{}{},
		Resp: &metadata,
	})
	if err != nil {
		c.router.Close()
	}
	c.siblings = metadata.Siblings
	return
}

// Closes and destroys the client's RPC server
func (c *Client) Disconnect() {
	if c.router == nil {
		return
	}
	c.router.Close()
	c.router = nil
}

// Get siblings so that each can be uploaded to individually.  This should be
// moved to a (c *Client) function that updates the current siblings. I'm
// actually considering that a client should listen on a quorum, or somehow
// perform lightweight actions (receive digests?) that allow it to keep up but
// don't require many resources.
func (c *Client) RetrieveSiblings() (err error) {
	// Iterate through known siblings until someone provides an updated list. The
	// first answer given is trusted, this is insecure.
	var metadata state.StateMetadata
	for i := range c.siblings {
		if c.siblings[i].Address.Host == "" {
			continue
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.siblings[i].Address,
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &metadata,
		})
		if err == nil {
			break
		}
		c.siblings = metadata.Siblings
	}
	if err != nil {
		err = errors.New("Could not reach any stored siblings")
		return
	}

	return
}

// This new function is a bit unique because it can return an error while also
// returning a fully working client.
func NewClient() (c *Client, err error) {
	c = new(Client)
	c.genericWallets = make(map[state.WalletID]*Keypair)
	err = c.Connect("localhost", 9988, 1) // default bootstrap address

	/* The format for config files
	directories:
		/path/to/a/directory/with/wallets/
		/another/wallet/path/
		/however/many/paths/you/want/
	wallet: 						<--This is optional. Only include if you want
		walletIDinhex				 to automatically load into a specific wallet
	*/
	f, err := os.Open(".config")
	r := bufio.NewReader(f)
	l, err := r.ReadString('\n')
	if strings.TrimSpace(l) != "directories:" {
		errors.New("Invalid config file")
		return
	}
	l, err = r.ReadString('\n')
	l = strings.TrimSpace(l)
	//Read in wallet directories and load wallets
	for l != "" {
		filenames, err := filepath.Glob(l + "*.id")
		if err != nil {
			panic(err)
		}
		for _, j := range filenames {
			id, keypair, err := LoadWallet(j)
			if err != nil {
				panic(err)
			}
			c.genericWallets[id] = keypair
		}
		l, err = r.ReadString('\n')
		l = strings.TrimSpace(l)
	}
	//Load starting wallet ID, if a starting wallet ID is desired
	l, err = r.ReadString('\n')
	if strings.TrimSpace(l) != "wallet:" {
		return
	}
	l, err = r.ReadString('\n')
	l = strings.TrimSpace(l)
	_, err = fmt.Sscanf(l, "%x", &c.CurID)
	return
}
