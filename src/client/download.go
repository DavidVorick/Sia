package client

import (
	"quorum"
)

// download every segment
// call repair
// look at the padding value
// truncate the padding

func (c *Client) Download(id quorum.WalletID, destination string) {
	c.RetrieveSiblings()

	// make files to store the incoming segments
	// file writers?
	//  what does the fetch function look like?
	//for i := range c.siblings {
	//}
}
