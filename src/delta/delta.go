// The delta layer manages inputs determined by the concensus layer and makes
// corresponding changes to the wallet layer.
package delta

import (
	"quorum"
)

type Engine struct {
	quorum Quorum
}
