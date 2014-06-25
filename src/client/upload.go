package client

import (
	"math"
	"os"
	"quorum"
)

func CalculateAtoms(filename string, m int) (atoms int, err error) {
	multiplier := float64(m) / float64(quorum.QuorumSize)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	info, err := file.Stat()
	if err != nil {
		return
	}
	size := info.Size()

	floatAtoms := multiplier * float64(size) / float64(quorum.AtomSize)
	atoms = int(math.Ceil(floatAtoms))
	return
}
