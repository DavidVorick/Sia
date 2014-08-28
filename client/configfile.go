package client

import (
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
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

// Iterates through the client and saves all of the wallets to disk.
func (c *Client) SaveAllWallets() (err error) {
	var filename string
	for id, keypair := range c.genericWallets {
		filename = fmt.Sprintf("%x.id", id)
		err = SaveWallet(state.WalletID(id), keypair, filename)
		if err != nil {
			return
		}
	}
	// Save other types of wallets as they are implemented.
	return
}

func SaveWallet(id state.WalletID, keypair GenericWallet, destFile string) (err error) {
	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(siaencoding.EncUint64(uint64(id)))
	if err != nil {
		return
	}
	_, err = f.Write(keypair.PublicKey[:])
	if err != nil {
		return
	}
	_, err = f.Write(keypair.SecretKey[:])
	if err != nil {
		return
	}
	return
}

func LoadWallet(fileName string) (id state.WalletID, keypair GenericWallet, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	idSlice := make([]byte, 8)
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(idSlice)
	if err != nil {
		return
	}
	_, err = f.Read(pubSlice)
	if err != nil {
		return
	}
	_, err = f.Read(secSlice)
	if err != nil {
		return
	}
	id = state.WalletID(siaencoding.DecUint64(idSlice))
	if copy(keypair.PublicKey[:], pubSlice) != siacrypto.PublicKeySize {
		err = errors.New("bad public key length")
		return
	}
	if copy(keypair.SecretKey[:], secSlice) != siacrypto.SecretKeySize {
		err = errors.New("bad secret key length")
		return
	}

	return
}
