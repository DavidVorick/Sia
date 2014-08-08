package client

import (
	"os"
	"os/user"
)

//The format for config files:
/*
directories:
	/path/to/a/directory/with/wallets/
	/another/wallet/path/
	/however/many/paths/you/want/
*/

// In the config file, the wallet variable is optional, and should only be used
// if you wanto to automatically load a specific wallet.

// processConfigFile opens, parses, and processes the config file, which will
// be stored in the home directory at .config/Sia/config
func (c *Client) processConfigFile() (err error) {
	// Get the filename of the config file, which will be stored in
	// $HOME/.config/Sia/config
	userObj, err := user.Current()
	if err != nil {
		return
	}
	filefolder := userObj.HomeDir + "/.config/Sia/"
	filename := filefolder + ".config"

	// Make the folder, in case it does not yet exist.
	err = os.MkdirAll(filefolder, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	// I'm commenting all of this out now because a generic parser needs to
	// be written. It seems like we've settled on YAML as our format of
	// choice.
	/*
		file, err := os.Open(".config")
		configReader := bufio.NewReader(file)
		l, err := configReader.ReadString('\n')
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
	*/
	return
}
